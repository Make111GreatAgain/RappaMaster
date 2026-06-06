#!/usr/bin/env python3
"""
ABM_V2 offline HTTP test client.

This script uses only the Python standard library so it can run on deployment
servers without internet access or pip installs.
"""

import argparse
import json
import sys
import time
from datetime import datetime
from pathlib import Path
from typing import Any, Dict, List, Optional
from urllib.error import HTTPError, URLError
from urllib.parse import urlencode
from urllib.request import Request, urlopen


DEFAULT_BASE_URL = "http://127.0.0.1:8081"
DEFAULT_STOCK_CODE = "600000"
DEFAULT_STOCK_NAME = "浦发银行"
DEFAULT_TIMEOUT = 60
CACHE_PATH = Path(__file__).resolve().with_name(".abm_task_http_client_cache.json")
REPORT_DIR = Path(__file__).resolve().parent / "test_reports"


class HttpClient:
    def __init__(self, base_url: str, timeout: int, report_path: Optional[Path] = None) -> None:
        self.base_url = base_url.rstrip("/")
        self.timeout = timeout
        self.report_path = report_path

    def get(self, path: str, query: Optional[Dict[str, Any]] = None) -> Dict[str, Any]:
        return self._request("GET", path, query=query)

    def post(self, path: str, query: Optional[Dict[str, Any]] = None, body: Any = None) -> Dict[str, Any]:
        return self._request("POST", path, query=query, body=body)

    def _request(
        self,
        method: str,
        path: str,
        query: Optional[Dict[str, Any]] = None,
        body: Any = None,
    ) -> Dict[str, Any]:
        url = f"{self.base_url}{path}"
        clean_query = {k: v for k, v in (query or {}).items() if v is not None and v != ""}
        if clean_query:
            url = f"{url}?{urlencode(clean_query)}"

        data = None
        headers = {"Content-Type": "application/json"}
        if body is not None:
            data = json.dumps(body, ensure_ascii=False).encode("utf-8")
        req = Request(url, data=data, method=method, headers=headers)

        started = time.time()
        status = None
        text = ""
        error = ""
        try:
            with urlopen(req, timeout=self.timeout) as resp:
                status = resp.status
                text = resp.read().decode("utf-8", errors="replace")
        except HTTPError as exc:
            status = exc.code
            text = exc.read().decode("utf-8", errors="replace")
            error = str(exc)
        except URLError as exc:
            error = str(exc)
        except Exception as exc:
            error = str(exc)

        elapsed_ms = int((time.time() - started) * 1000)
        parsed = parse_json(text)
        result = {
            "method": method,
            "url": url,
            "status": status,
            "elapsedMs": elapsed_ms,
            "requestBody": body,
            "responseJson": parsed,
            "responseText": text,
            "error": error,
        }
        print_response(result)
        self._append_report(result)
        return result

    def _append_report(self, item: Dict[str, Any]) -> None:
        if not self.report_path:
            return
        with open(self.report_path, "a", encoding="utf-8") as f:
            f.write(json.dumps(item, ensure_ascii=False, indent=2))
            f.write("\n\n")


def parse_json(text: str) -> Optional[Any]:
    if not text:
        return None
    try:
        return json.loads(text)
    except Exception:
        return None


def print_response(item: Dict[str, Any]) -> None:
    print(f"\n{item['method']} {item['url']}")
    print(f"status={item['status']} elapsedMs={item['elapsedMs']}")
    if item["error"]:
        print(f"error={item['error']}")
    if item["responseJson"] is not None:
        print(json.dumps(item["responseJson"], ensure_ascii=False, indent=2))
    elif item["responseText"]:
        print(item["responseText"][:4000])


def load_cache() -> Dict[str, Any]:
    if not CACHE_PATH.exists():
        return {}
    try:
        return json.loads(CACHE_PATH.read_text(encoding="utf-8"))
    except Exception:
        return {}


def save_cache(cache: Dict[str, Any]) -> None:
    CACHE_PATH.write_text(json.dumps(cache, ensure_ascii=False, indent=2), encoding="utf-8")


def cached_task_id() -> str:
    return str(load_cache().get("taskId") or "")


def cache_task_id(task_id: str) -> None:
    if not task_id:
        return
    cache = load_cache()
    cache["taskId"] = task_id
    save_cache(cache)
    print(f"cached taskId={task_id}")


def extract_status(payload: Any) -> str:
    if not isinstance(payload, dict):
        return ""
    return str(payload.get("status") or payload.get("code") or "")


def extract_data(payload: Any) -> Any:
    if not isinstance(payload, dict):
        return None
    return payload.get("data")


def extract_task_id_from_create(result: Dict[str, Any]) -> str:
    data = extract_data(result.get("responseJson"))
    if isinstance(data, dict):
        return str(data.get("taskId") or "")
    return ""


def build_immediate_payload(stock_code: str, stock_name: str, data_start: str = "", data_end: str = "") -> List[Dict[str, Any]]:
    item: Dict[str, Any] = {
        "stockCode": stock_code,
        "stockName": stock_name or stock_code,
        "N_FT": 20,
        "N_LMT": 20,
        "N_SMT": 20,
        "N_NT": 20,
        "ALPHA_L": 0.001,
        "ALPHA_S": 0.9,
        "S_FT": 1,
        "fundamental_value": "XX",
        "horizon": "1天 (T+1)",
    }
    if data_start:
        item["dataStartDate"] = data_start
    if data_end:
        item["dataEndDate"] = data_end
    return [item]


def create_immediate(client: HttpClient, args: argparse.Namespace) -> str:
    result = client.post(
        "/simulation/create-task",
        query={"isScheduled": "false"},
        body=build_immediate_payload(args.stock_code, args.stock_name, args.data_start, args.data_end),
    )
    task_id = extract_task_id_from_create(result)
    cache_task_id(task_id)
    return task_id


def create_scheduled(client: HttpClient, args: argparse.Namespace) -> str:
    result = client.post(
        "/simulation/create-task",
        query={
            "isScheduled": "true",
            "universe": args.universe,
            "dataStartDate": args.data_start,
            "dataEndDate": args.data_end,
        },
        body=None,
    )
    task_id = extract_task_id_from_create(result)
    cache_task_id(task_id)
    return task_id


def query_abm_parameters(client: HttpClient, stock_code: str = "") -> Dict[str, Any]:
    return client.get("/simulation/abm_parameters", {"stockCode": stock_code})


def refresh_abm_parameters(client: HttpClient) -> Dict[str, Any]:
    return client.post("/internal/simulation/abm_parameters/refresh")


def query_execution_log(client: HttpClient) -> Dict[str, Any]:
    return client.get("/dashboard/execution_log")


def query_task_detail(client: HttpClient, task_id: str) -> Dict[str, Any]:
    return client.get("/dashboard/execution_log/task", {"taskId": task_id})


def task_finished_from_detail(result: Dict[str, Any]) -> bool:
    payload = result.get("responseJson")
    data = extract_data(payload)
    text = json.dumps(data, ensure_ascii=False).lower() if data is not None else ""
    return any(word in text for word in ["finished", "success", "completed", "完成"])


def collect_stock_ids_from_detail(result: Dict[str, Any], fallback_stock: str) -> List[str]:
    payload = result.get("responseJson")
    data = extract_data(payload)
    text = json.dumps(data, ensure_ascii=False)
    candidates = sorted(set(part for part in split_non_digits(text) if len(part) == 6 and part.isdigit()))
    if fallback_stock and fallback_stock not in candidates:
        candidates.insert(0, fallback_stock)
    return candidates[:20]


def split_non_digits(text: str) -> List[str]:
    current = []
    out = []
    for ch in text:
        if ch.isdigit():
            current.append(ch)
        elif current:
            out.append("".join(current))
            current = []
    if current:
        out.append("".join(current))
    return out


def poll_task(client: HttpClient, task_id: str, interval: int, max_wait: int) -> Dict[str, Any]:
    deadline = time.time() + max_wait
    last = {}
    while True:
        last = query_task_detail(client, task_id)
        if task_finished_from_detail(last):
            print(f"task {task_id} looks finished")
            return last
        if time.time() >= deadline:
            print(f"poll timeout after {max_wait}s")
            return last
        time.sleep(interval)


def query_analytics(client: HttpClient, task_id: str, stock_id: str) -> None:
    query = {"taskId": task_id, "stockId": stock_id}
    client.get("/dashboard/order_dynamics", query)
    client.get("/dashboard/price_synthesis", query)
    client.get("/dashboard/crash_risk_warning", query)
    client.get("/dashboard/investor_composition", query)
    client.get("/dashboard/performance_comparison", query)
    client.get("/dashboard/performance_comparison", {**query, "selectedModel": "ABM"})
    client.get("/dashboard/performance_comparison", {**query, "selectedModel": "VRNN"})
    client.get("/dashboard/performance_comparison", {**query, "selectedModel": "TimeGAN"})


def smoke(client: HttpClient, args: argparse.Namespace) -> None:
    query_abm_parameters(client)
    query_abm_parameters(client, args.stock_code)
    query_execution_log(client)


def run_full(client: HttpClient, args: argparse.Namespace) -> None:
    smoke(client, args)
    task_id = create_scheduled(client, args) if args.scheduled else create_immediate(client, args)
    if not task_id:
        print("create task did not return taskId; stop full flow")
        return
    detail = poll_task(client, task_id, args.poll_interval, args.max_wait)
    stock_ids = collect_stock_ids_from_detail(detail, args.stock_code)
    if not stock_ids:
        print("no stockId detected for analytics query")
        return
    query_analytics(client, task_id, stock_ids[0])


def make_report_path(command: str) -> Path:
    REPORT_DIR.mkdir(parents=True, exist_ok=True)
    ts = datetime.now().strftime("%Y%m%d_%H%M%S")
    return REPORT_DIR / f"abm_{command}_{ts}.log"


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(description="Offline ABM_V2 HTTP test client")
    parser.add_argument("--base-url", default=DEFAULT_BASE_URL)
    parser.add_argument("--timeout", type=int, default=DEFAULT_TIMEOUT)
    parser.add_argument("--stock-code", default=DEFAULT_STOCK_CODE)
    parser.add_argument("--stock-name", default=DEFAULT_STOCK_NAME)
    parser.add_argument("--task-id", default="")
    parser.add_argument("--universe", default="all", choices=["all", "hs300", "csi1000"])
    parser.add_argument("--data-start", default="")
    parser.add_argument("--data-end", default="")
    parser.add_argument("--poll-interval", type=int, default=30)
    parser.add_argument("--max-wait", type=int, default=3600)
    parser.add_argument("--no-report", action="store_true")
    return parser


def main() -> int:
    parser = build_parser()
    args = parser.parse_args()
    report_path = None if args.no_report else make_report_path("shell")
    if report_path:
        print(f"report: {report_path}")
    client = HttpClient(args.base_url, args.timeout, report_path)
    interactive_shell(client, args)
    return 0


def require_task_id(task_id: str) -> None:
    if not task_id:
        print("taskId is required. Pass --task-id or create a task first to populate cache.", file=sys.stderr)
        raise SystemExit(2)


def interactive_shell(client: HttpClient, args: argparse.Namespace) -> None:
    print("ABM_V2 offline HTTP client shell")
    print("Commands:")
    print("  smoke")
    print("  refresh")
    print("  params [stockCode]")
    print("  create_immediate [stockCode] [stockName] [dataStartDate] [dataEndDate]")
    print("  create_scheduled [all|hs300|csi1000] [dataStartDate] [dataEndDate]")
    print("  poll [taskId]")
    print("  analytics [taskId] [stockCode]")
    print("  log")
    print("  help")
    print("  exit")

    while True:
        try:
            raw = input("abm> ").strip()
        except (EOFError, KeyboardInterrupt):
            print()
            return
        if not raw:
            continue
        parts = raw.split()
        command = parts[0].lower()
        if command in {"exit", "quit", "q"}:
            return
        if command in {"help", "h", "?"}:
            print_shell_help()
            continue
        try:
            if command == "smoke":
                smoke(client, args)
            elif command in {"refresh", "refresh_params"}:
                refresh_abm_parameters(client)
            elif command == "params":
                stock_code = parts[1] if len(parts) > 1 else ""
                query_abm_parameters(client, stock_code)
            elif command == "create_immediate":
                stock_code = parts[1] if len(parts) > 1 else prompt_default("stockCode", args.stock_code)
                stock_name = parts[2] if len(parts) > 2 else prompt_default("stockName", args.stock_name or stock_code)
                data_start = parts[3] if len(parts) > 3 else prompt_default("dataStartDate", "")
                data_end = parts[4] if len(parts) > 4 else prompt_default("dataEndDate", data_start)
                shell_args = clone_args(args, stock_code=stock_code, stock_name=stock_name, data_start=data_start, data_end=data_end)
                create_immediate(client, shell_args)
            elif command == "create_scheduled":
                universe = parts[1] if len(parts) > 1 else prompt_default("universe(all/hs300/csi1000)", args.universe)
                data_start = parts[2] if len(parts) > 2 else prompt_default("dataStartDate", "")
                data_end = parts[3] if len(parts) > 3 else prompt_default("dataEndDate", data_start)
                shell_args = clone_args(args, universe=universe, data_start=data_start, data_end=data_end)
                create_scheduled(client, shell_args)
            elif command == "poll":
                task_id = parts[1] if len(parts) > 1 else cached_task_id()
                require_task_id(task_id)
                poll_task(client, task_id, args.poll_interval, args.max_wait)
            elif command == "analytics":
                task_id = parts[1] if len(parts) > 1 else cached_task_id()
                stock_code = parts[2] if len(parts) > 2 else prompt_default("stockCode", args.stock_code)
                require_task_id(task_id)
                query_analytics(client, task_id, stock_code)
            elif command == "log":
                query_execution_log(client)
            else:
                print(f"unknown command: {command}")
                print_shell_help()
        except SystemExit:
            raise
        except Exception as exc:
            print(f"command failed: {exc}")


def print_shell_help() -> None:
    print("Examples:")
    print("  params 600000")
    print("  create_immediate 600000 浦发银行 2026-01-05 2026-01-05")
    print("  create_scheduled hs300")
    print("  create_scheduled csi1000 2026-01-05 2026-01-09")
    print("  poll")
    print("  analytics TSK-1001 600000")


def prompt_default(label: str, default: str) -> str:
    suffix = f" [{default}]" if default else ""
    value = input(f"{label}{suffix}: ").strip()
    return value if value else default


def clone_args(args: argparse.Namespace, **overrides: Any) -> argparse.Namespace:
    data = vars(args).copy()
    data.update(overrides)
    return argparse.Namespace(**data)


if __name__ == "__main__":
    raise SystemExit(main())
