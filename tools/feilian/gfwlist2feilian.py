import base64
import json
import re
from pathlib import Path
from urllib.parse import urlparse
import ipaddress


INPUT_FILE = "gfwlist.b64"
OUTPUT_PREFIX = "feilian_gfwlist"
MAX_CHARS = 65000  # 飞连限制 65536，这里留一点余量


def decode_gfwlist(path: str) -> str:
    raw = Path(path).read_bytes()
    raw = b"".join(raw.split())

    # 兼容缺少 padding 的 base64
    missing_padding = len(raw) % 4
    if missing_padding:
        raw += b"=" * (4 - missing_padding)

    return base64.b64decode(raw).decode("utf-8", errors="ignore")


def clean_domain(domain: str) -> str | None:
    domain = domain.strip().lower()

    domain = domain.strip("|")
    domain = domain.strip("^")
    domain = domain.strip("/")
    domain = domain.strip(".")

    if not domain:
        return None

    # 去掉端口
    if ":" in domain and not re.match(r"^\[.*\]$", domain):
        domain = domain.split(":", 1)[0]

    # 去掉通配符前缀，后面统一补裸域名 + 泛域名
    if domain.startswith("*."):
        domain = domain[2:]

    # 过滤明显不是域名的内容
    if "/" in domain or "*" in domain or "@" in domain:
        return None

    if not re.match(r"^[a-z0-9.-]+$", domain):
        return None

    if "." not in domain:
        return None

    if domain.startswith("-") or domain.endswith("-"):
        return None

    return domain


def extract_host_from_rule(rule: str) -> str | None:
    rule = rule.strip()

    # 跳过白名单规则
    if rule.startswith("@@"):
        return None

    # 跳过注释和元信息
    if not rule or rule.startswith("!") or rule.startswith("["):
        return None

    # 跳过元素隐藏规则
    if "##" in rule or "#@#" in rule or "#?#" in rule:
        return None

    # 去掉 ABP option，例如 ||example.com^$third-party
    if "$" in rule:
        rule = rule.split("$", 1)[0]

    # 形如 ||example.com^
    if rule.startswith("||"):
        body = rule[2:]
        body = re.split(r"[\^/*]", body, maxsplit=1)[0]
        return clean_domain(body)

    # 形如 .example.com
    if rule.startswith("."):
        return clean_domain(rule)

    # 形如 |http://example.com/path
    if rule.startswith("|http://") or rule.startswith("|https://"):
        rule = rule[1:]

    # 形如 http://example.com/path
    if rule.startswith("http://") or rule.startswith("https://"):
        parsed = urlparse(rule)
        return clean_domain(parsed.hostname or "")

    # 形如 example.com/path
    if "/" in rule:
        first = rule.split("/", 1)[0]
        return clean_domain(first)

    # 形如 example.com
    return clean_domain(rule)


def domain_sort_key(domain: str) -> tuple[str, int]:
    """
    排序时按照去掉 *. 后的域名排序。
    同一个主域名下，裸域名排在泛域名前面。

    example:
    google.com
    *.google.com
    mail.google.com
    *.mail.google.com
    """
    if domain.startswith("*."):
        return domain[2:], 1
    return domain, 0


def build_resource_json(name: str, domains: list[str], ips: list[str]) -> str:
    data = [
        {
            "name": name,
            "ip": ips,
            "dynamic_domain": domains,
            "static_domain": []
        }
    ]

    return json.dumps(data, ensure_ascii=False, indent=2)


def split_domain_groups(
    base_domains: list[str],
    ips: list[str],
    max_chars: int
) -> list[tuple[list[str], list[str]]]:
    """
    按主域名拆分，保证 example.com 和 *.example.com 不会被拆到两个文件里。
    IP 只放在第一个文件里。
    """
    chunks = []
    current_domains = []
    first_chunk = True

    for base_domain in base_domains:
        group = [
            base_domain,
            f"*.{base_domain}"
        ]

        trial_domains = current_domains + group
        trial_ips = ips if first_chunk else []

        trial_json = build_resource_json(
            name=f"gfwlist-{len(chunks) + 1:03d}",
            domains=trial_domains,
            ips=trial_ips
        )

        if len(trial_json) <= max_chars:
            current_domains = trial_domains
            continue

        if not current_domains:
            raise ValueError(f"单个域名组已经超过限制: {group}")

        chunks.append((current_domains, trial_ips))
        first_chunk = False
        current_domains = group

        retry_json = build_resource_json(
            name=f"gfwlist-{len(chunks) + 1:03d}",
            domains=current_domains,
            ips=[]
        )

        if len(retry_json) > max_chars:
            raise ValueError(f"单个域名组已经超过限制: {group}")

    if current_domains:
        chunks.append((current_domains, ips if first_chunk else []))

    return chunks


def main():
    text = decode_gfwlist(INPUT_FILE)

    base_domains = set()
    ips = set()

    for line in text.splitlines():
        line = line.strip()

        if not line:
            continue

        # IP 或 CIDR
        try:
            ip_obj = ipaddress.ip_network(line, strict=False)
            ips.add(str(ip_obj))
            continue
        except ValueError:
            pass

        domain = extract_host_from_rule(line)
        if domain:
            base_domains.add(domain)

    sorted_base_domains = sorted(base_domains)
    sorted_ips = sorted(ips)

    chunks = split_domain_groups(
        base_domains=sorted_base_domains,
        ips=sorted_ips,
        max_chars=MAX_CHARS
    )

    for index, (domains, chunk_ips) in enumerate(chunks, start=1):
        name = f"gfwlist-{index:03d}"
        output_json = build_resource_json(
            name=name,
            domains=domains,
            ips=chunk_ips
        )

        output_file = f"{OUTPUT_PREFIX}_{index:03d}.json"
        Path(output_file).write_text(output_json, encoding="utf-8")

        print(
            f"{output_file}: "
            f"{len(output_json)} chars, "
            f"{len(domains)} domains, "
            f"{len(chunk_ips)} ips"
        )

    print()
    print(f"总主域名数量: {len(sorted_base_domains)}")
    print(f"总 dynamic_domain 数量: {len(sorted_base_domains) * 2}")
    print(f"总 IP 数量: {len(sorted_ips)}")
    print(f"生成文件数量: {len(chunks)}")


if __name__ == "__main__":
    main()
