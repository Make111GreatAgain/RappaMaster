#!/usr/bin/env python3
"""Refresh A-share stock code/name mapping for Master HTTP responses.

The generated file is a plain JSON object:

{
  "600000": "浦发银行",
  "000001": "平安银行"
}

It intentionally avoids third-party dependencies so it can run in the
deployment environment without installing AkShare/Pandas.
"""

import argparse
import json
import pathlib
import sys
import time
import urllib.parse
import urllib.request


EASTMONEY_URL = "https://push2.eastmoney.com/api/qt/clist/get"
EASTMONEY_FS = "m:0+t:6,m:0+t:80,m:1+t:2,m:1+t:23,m:0+t:81+s:2048"


def fetch_json(url: str, params: dict, timeout: int) -> dict:
    req = urllib.request.Request(
        url + "?" + urllib.parse.urlencode(params),
        headers={
            "User-Agent": "Mozilla/5.0",
            "Referer": "https://quote.eastmoney.com/center/gridlist.html",
        },
    )
    with urllib.request.urlopen(req, timeout=timeout) as resp:
        return json.loads(resp.read().decode("utf-8"))


def fetch_eastmoney_mapping(page_size: int, timeout: int, sleep_seconds: float) -> dict:
    stock_map = {}
    page = 1
    total = None

    while True:
        payload = fetch_json(
            EASTMONEY_URL,
            {
                "pn": page,
                "pz": page_size,
                "po": 1,
                "np": 1,
                "ut": "bd1d9ddb04089700cf9c27f6f7426281",
                "fltt": 2,
                "invt": 2,
                "fid": "f3",
                "fs": EASTMONEY_FS,
                "fields": "f12,f14",
            },
            timeout,
        )
        if payload.get("rc") != 0:
            raise RuntimeError(f"eastmoney returned rc={payload.get('rc')}: {payload}")

        data = payload.get("data") or {}
        total = data.get("total", total)
        rows = data.get("diff") or []
        if not rows:
            break

        for row in rows:
            code = str(row.get("f12", "")).strip()
            name = str(row.get("f14", "")).strip()
            if len(code) == 6 and code.isdigit() and name:
                stock_map[code] = name

        if total is not None and len(stock_map) >= int(total):
            break
        page += 1
        if sleep_seconds > 0:
            time.sleep(sleep_seconds)

    return dict(sorted(stock_map.items()))


def main() -> int:
    parser = argparse.ArgumentParser(description="Update Config/stock_name_map.json")
    parser.add_argument(
        "--output",
        default="/root/rappa/RappaMaster/Config/stock_name_map.json",
        help="Output JSON path",
    )
    parser.add_argument("--page-size", type=int, default=1000)
    parser.add_argument("--timeout", type=int, default=20)
    parser.add_argument("--sleep", type=float, default=0.1)
    args = parser.parse_args()

    stock_map = fetch_eastmoney_mapping(args.page_size, args.timeout, args.sleep)
    if not stock_map:
        raise RuntimeError("empty stock mapping")

    output = pathlib.Path(args.output)
    output.parent.mkdir(parents=True, exist_ok=True)
    output.write_text(
        json.dumps(stock_map, ensure_ascii=False, indent=2) + "\n",
        encoding="utf-8",
    )
    print(f"wrote {len(stock_map)} stock names to {output}")
    return 0


if __name__ == "__main__":
    try:
        raise SystemExit(main())
    except Exception as exc:
        print(f"update stock name map failed: {exc}", file=sys.stderr)
        raise
