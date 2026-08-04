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
    parser.add_argument("--stock-code", default="")
    parser.add_argument("--stock-codes", default="")
    parser.add_argument("--start-date", required=True)
    parser.add_argument("--end-date", required=True)
    return parser.parse_args()


def normalize_stock_code(raw: str) -> str:
    stock_code = "".join(ch for ch in str(raw) if ch.isdigit())
    if len(stock_code) < 6:
        stock_code = stock_code.zfill(6)
    return stock_code


def parse_stock_codes(args: argparse.Namespace) -> list[str]:
    raw_values = []
    if args.stock_codes:
        raw_values.extend(args.stock_codes.split(","))
    if args.stock_code:
        raw_values.append(args.stock_code)

    result = []
    seen = set()
    for raw in raw_values:
        code = normalize_stock_code(raw)
        if not code or code in seen:
            continue
        seen.add(code)
        result.append(code)
    return result


def check_one_stock(session, args: argparse.Namespace, stock_code: str) -> dict:
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
    result = session.run(script)
    rows = 0
    if hasattr(result, "iloc"):
        rows = int(result.iloc[0, 0])
    elif isinstance(result, list) and result:
        rows = int(result[0])
    else:
        rows = int(result)
    return {"exists": rows > 0, "rows": rows, "reason": "" if rows > 0 else "no remote rows"}


def main() -> int:
    args = parse_args()
    try:
        import dolphindb as ddb
    except Exception as exc:
        print(json.dumps({"exists": False, "rows": 0, "reason": f"import dolphindb failed: {exc}"}, ensure_ascii=False))
        return 2

    stock_codes = parse_stock_codes(args)
    if not stock_codes:
        print(json.dumps({"exists": False, "rows": 0, "reason": "stock-code is required"}, ensure_ascii=False))
        return 2

    try:
        session = ddb.session()
        session.connect(host=args.host, port=args.port, userid=args.user, password=args.password, keepAliveTime=300)
        stocks = {stock_code: check_one_stock(session, args, stock_code) for stock_code in stock_codes}
        if len(stock_codes) == 1:
            single = stocks[stock_codes[0]]
            print(json.dumps(single, ensure_ascii=False))
            return 0
        print(json.dumps({"stocks": stocks}, ensure_ascii=False))
        return 0
    except Exception as exc:
        if len(stock_codes) == 1:
            print(json.dumps({"exists": False, "rows": 0, "reason": str(exc)}, ensure_ascii=False))
            return 3
        print(json.dumps({
            "stocks": {
                stock_code: {"exists": False, "rows": 0, "reason": str(exc)}
                for stock_code in stock_codes
            }
        }, ensure_ascii=False))
        return 3


if __name__ == "__main__":
    sys.exit(main())
