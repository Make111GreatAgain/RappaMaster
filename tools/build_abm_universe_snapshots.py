#!/usr/bin/env python3
from __future__ import annotations

import argparse
import csv
import re
import xml.etree.ElementTree as ET
from dataclasses import dataclass
from datetime import date, datetime
from pathlib import Path
from zipfile import ZipFile

NS = {
    "a": "http://schemas.openxmlformats.org/spreadsheetml/2006/main",
    "r": "http://schemas.openxmlformats.org/officeDocument/2006/relationships",
}
REL_NS = {"rel": "http://schemas.openxmlformats.org/package/2006/relationships"}


@dataclass
class Adjustment:
    effective_date: date
    stock_code: str
    stock_name: str
    action: str


def column_number(cell_ref: str) -> int:
    n = 0
    for ch in "".join(c for c in cell_ref if c.isalpha()):
        n = n * 26 + ord(ch.upper()) - 64
    return n


def normalize_code(raw: str) -> str:
    digits = "".join(ch for ch in str(raw or "") if ch.isdigit())
    if not digits:
        return ""
    return digits.zfill(6)[-6:]


def parse_date(raw: str) -> date:
    text = str(raw or "").replace("\xa0", "").strip()
    return datetime.strptime(text, "%Y-%m-%d").date()


def read_xlsx_rows(path: Path) -> list[list[str]]:
    with ZipFile(path) as z:
        shared = []
        try:
            root = ET.fromstring(z.read("xl/sharedStrings.xml"))
            for item in root.findall("a:si", NS):
                shared.append("".join(t.text or "" for t in item.findall(".//a:t", NS)))
        except KeyError:
            pass

        workbook = ET.fromstring(z.read("xl/workbook.xml"))
        rels = ET.fromstring(z.read("xl/_rels/workbook.xml.rels"))
        rel_map = {rel.attrib["Id"]: rel.attrib["Target"] for rel in rels.findall("rel:Relationship", REL_NS)}
        sheet = workbook.findall(".//a:sheet", NS)[0]
        rel_id = sheet.attrib["{http://schemas.openxmlformats.org/officeDocument/2006/relationships}id"]
        sheet_path = "xl/" + rel_map[rel_id].lstrip("/")

        def cell_value(cell: ET.Element) -> str:
            value = cell.find("a:v", NS)
            cell_type = cell.attrib.get("t")
            if cell_type == "inlineStr":
                return "".join(t.text or "" for t in cell.findall(".//a:t", NS))
            if value is None:
                return ""
            raw_value = value.text or ""
            if cell_type == "s" and raw_value.isdigit():
                return shared[int(raw_value)]
            return raw_value

        sheet_root = ET.fromstring(z.read(sheet_path))
        rows = []
        for row in sheet_root.findall(".//a:sheetData/a:row", NS):
            cells = {column_number(cell.attrib.get("r", "")): cell_value(cell) for cell in row.findall("a:c", NS)}
            if cells:
                rows.append([cells.get(i, "") for i in range(1, max(cells) + 1)])
        return rows


def load_adjustments(path: Path, kind: str) -> list[Adjustment]:
    rows = read_xlsx_rows(path)
    header = rows[0]
    if kind == "hs300":
        date_idx = header.index("变更日期")
        code_idx = header.index("成份证券代码")
        name_idx = header.index("成份证券简称")
        action_idx = header.index("变动方式")
        in_words = {"调入"}
    else:
        date_idx = header.index("日期")
        code_idx = header.index("代码")
        name_idx = header.index("简称")
        action_idx = header.index("纳入/剔除")
        in_words = {"纳入"}

    adjustments = []
    for row in rows[1:]:
        if len(row) <= max(date_idx, code_idx, name_idx, action_idx):
            continue
        code = normalize_code(row[code_idx])
        if not code:
            continue
        action = "in" if str(row[action_idx]).strip() in in_words else "out"
        adjustments.append(Adjustment(parse_date(row[date_idx]), code, str(row[name_idx]).strip(), action))
    adjustments.sort(key=lambda item: (item.effective_date, item.stock_code, item.action))
    return adjustments


def quarter_end(year: int, quarter: int) -> date:
    month_day = {1: (3, 31), 2: (6, 30), 3: (9, 30), 4: (12, 31)}[quarter]
    return date(year, month_day[0], month_day[1])


def quarter_label(day: date) -> str:
    return f"{day.year}Q{((day.month - 1) // 3) + 1}"


def next_quarter(year: int, quarter: int) -> tuple[int, int]:
    if quarter == 4:
        return year + 1, 1
    return year, quarter + 1


def build_snapshots(adjustments: list[Adjustment], output_dir: Path, start: date) -> None:
    output_dir.mkdir(parents=True, exist_ok=True)
    latest = max(item.effective_date for item in adjustments)
    year, quarter = start.year, ((start.month - 1) // 3) + 1
    current: dict[str, str] = {}
    index = 0

    while quarter_end(year, quarter) <= quarter_end(latest.year, ((latest.month - 1) // 3) + 1):
        end = quarter_end(year, quarter)
        while index < len(adjustments) and adjustments[index].effective_date <= end:
            item = adjustments[index]
            if item.action == "in":
                current[item.stock_code] = item.stock_name
            else:
                current.pop(item.stock_code, None)
            index += 1

        snapshot_path = output_dir / f"{year}Q{quarter}.csv"
        with open(snapshot_path, "w", encoding="utf-8", newline="") as f:
            writer = csv.writer(f)
            writer.writerow(["stockCode", "stockName"])
            for code in sorted(current):
                writer.writerow([code, current[code]])
        year, quarter = next_quarter(year, quarter)


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Build ABM universe quarterly snapshots from index adjustment xlsx files.")
    parser.add_argument("--hs300-xlsx", default="/root/rappa/codex_plan/沪深300指数成分股历年调整名单（2005-2025年）+调入和调出.xlsx")
    parser.add_argument("--csi1000-xlsx", default="/root/rappa/codex_plan/中证1000-成分进出记录-截止到20260205.xlsx")
    parser.add_argument("--output-root", default="resources/abm_universe")
    parser.add_argument("--start-quarter", default="2022Q1")
    return parser.parse_args()


def main() -> int:
    args = parse_args()
    match = re.fullmatch(r"(\d{4})Q([1-4])", args.start_quarter.strip().upper())
    if not match:
        raise ValueError("start-quarter must look like 2022Q1")
    start = date(int(match.group(1)), (int(match.group(2)) - 1) * 3 + 1, 1)
    output_root = Path(args.output_root)
    build_snapshots(load_adjustments(Path(args.hs300_xlsx), "hs300"), output_root / "hs300", start)
    build_snapshots(load_adjustments(Path(args.csi1000_xlsx), "csi1000"), output_root / "csi1000", start)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
