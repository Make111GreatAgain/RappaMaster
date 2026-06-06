#!/usr/bin/env python3
import argparse
import json
import sys


def dolphin_date(value: str) -> str:
    return value.strip().replace("-", ".")


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Check ABM DolphinDB stock minute data availability")
    parser.add_argument("--host", required=True)
    parser.add_argument("--port", required=True, type=int)
    parser.add_argument("--user", required=True)
    parser.add_argument("--password", required=True)
    parser.add_argument("--db", required=True)
    parser.add_argument("--table", required=True)
    parser.add_argument("--stock-code", required=True)
    parser.add_argument("--start-date", required=True)
    parser.add_argument("--end-date", required=True)
    return parser.parse_args()


def main() -> int:
    args = parse_args()
    try:
        import dolphindb as ddb
    except Exception as exc:
        print(json.dumps({"exists": False, "rows": 0, "reason": f"import dolphindb failed: {exc}"}, ensure_ascii=False))
        return 2

    stock_code = "".join(ch for ch in str(args.stock_code) if ch.isdigit())
    if len(stock_code) < 6:
        stock_code = stock_code.zfill(6)
    symbol = f"`{stock_code}"
    start = dolphin_date(args.start_date)
    end = dolphin_date(args.end_date)

    script = f"""
inputDBName = "{args.db}"
inputTBName = "{args.table}"
quotes = loadTable(inputDBName, inputTBName)
d = select top 1 Symbol from quotes where Symbol={symbol} and TradingDate between {start}:{end}
select count(*) as rows from d
"""
    try:
        session = ddb.session()
        session.connect(host=args.host, port=args.port, userid=args.user, password=args.password, keepAliveTime=300)
        result = session.run(script)
        rows = 0
        if hasattr(result, "iloc"):
            rows = int(result.iloc[0, 0])
        elif isinstance(result, list) and result:
            rows = int(result[0])
        else:
            rows = int(result)
        print(json.dumps({"exists": rows > 0, "rows": rows, "reason": "" if rows > 0 else "no remote rows"}, ensure_ascii=False))
        return 0
    except Exception as exc:
        print(json.dumps({"exists": False, "rows": 0, "reason": str(exc)}, ensure_ascii=False))
        return 3


if __name__ == "__main__":
    sys.exit(main())
