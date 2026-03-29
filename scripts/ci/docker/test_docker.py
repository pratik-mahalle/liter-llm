#!/usr/bin/env python3
"""Docker integration test suite for liter-llm proxy server.

Runs a series of tests against a Docker image to verify:
- CLI version output
- Health endpoints (liveness, readiness, full health)
- OpenAPI spec serving
- Authentication (missing/invalid/valid keys)
- Security (non-root user)

Usage:
    python3 scripts/ci/docker/test_docker.py --image liter-llm:test
    python3 scripts/ci/docker/test_docker.py  # defaults to liter-llm:test
"""

from __future__ import annotations

import argparse
import json
import subprocess
import sys
import time
from dataclasses import dataclass, field
from typing import Any

GREEN = "\033[92m"
RED = "\033[91m"
CYAN = "\033[96m"
RESET = "\033[0m"

HTTP_OK = 200
HTTP_UNAUTHORIZED = 401
HTTP_SERVICE_UNAVAILABLE = 503


@dataclass
class Results:
    total: int = 0
    passed: int = 0
    failed: int = 0
    failed_tests: list[str] = field(default_factory=list)

    def record(self, name: str, *, ok: bool) -> None:
        self.total += 1
        if ok:
            self.passed += 1
            print(f"  {GREEN}✓{RESET} {name}")
        else:
            self.failed += 1
            self.failed_tests.append(name)
            print(f"  {RED}✗{RESET} {name}")

    def summary(self) -> dict[str, Any]:
        return {
            "total_tests": self.total,
            "passed": self.passed,
            "failed": self.failed,
            "success_rate": round(self.passed / self.total * 100, 1) if self.total else 0,
            "failed_tests": self.failed_tests,
        }


def run(cmd: list[str], *, timeout: int = 30) -> subprocess.CompletedProcess[str]:
    return subprocess.run(cmd, capture_output=True, text=True, timeout=timeout, check=False)


def docker_run(image: str, *args: str, timeout: int = 30) -> subprocess.CompletedProcess[str]:
    return run(["docker", "run", "--rm", image, *args], timeout=timeout)


def curl(url: str, *, headers: dict[str, str] | None = None, timeout: int = 10) -> subprocess.CompletedProcess[str]:
    cmd = ["curl", "-sf", "--max-time", str(timeout)]
    for k, v in (headers or {}).items():
        cmd.extend(["-H", f"{k}: {v}"])
    cmd.append(url)
    return run(cmd, timeout=timeout + 5)


def curl_status(
    url: str,
    *,
    method: str = "GET",
    headers: dict[str, str] | None = None,
    data: str | None = None,
) -> int:
    """Return HTTP status code (0 on connection failure)."""
    cmd = ["curl", "-s", "-o", "/dev/null", "-w", "%{http_code}", "--max-time", "10", "-X", method]
    for k, v in (headers or {}).items():
        cmd.extend(["-H", f"{k}: {v}"])
    if data:
        cmd.extend(["-d", data])
    cmd.append(url)
    r = run(cmd, timeout=15)
    try:
        return int(r.stdout.strip())
    except ValueError:
        return 0


def start_container(image: str, port: int, *, env: dict[str, str] | None = None) -> str:
    """Start a detached container, return container ID."""
    cmd = ["docker", "run", "-d", "-p", f"{port}:4000"]
    for k, v in (env or {}).items():
        cmd.extend(["-e", f"{k}={v}"])
    cmd.append(image)
    r = run(cmd)
    return r.stdout.strip()


def stop_container(cid: str) -> None:
    run(["docker", "stop", cid])


def wait_healthy(port: int, *, retries: int = 15, delay: float = 1.0) -> bool:
    """Wait for /health/liveness to return 200."""
    for _ in range(retries):
        try:
            r = curl(f"http://localhost:{port}/health/liveness")
            if r.returncode == 0:
                return True
        except subprocess.TimeoutExpired:
            pass
        time.sleep(delay)
    return False


def test_version(image: str, results: Results) -> None:
    r = docker_run(image, "--version")
    ok = r.returncode == 0 and "liter-llm" in r.stdout
    results.record("CLI --version", ok=ok)
    if not ok:
        print(f"    stdout: {r.stdout!r}")


def test_help(image: str, results: Results) -> None:
    r = docker_run(image, "--help")
    ok = r.returncode == 0 and "api" in r.stdout and "mcp" in r.stdout
    results.record("CLI --help shows api and mcp commands", ok=ok)


def test_liveness(port: int, results: Results) -> None:
    status = curl_status(f"http://localhost:{port}/health/liveness")
    results.record("GET /health/liveness returns 200", ok=status == HTTP_OK)


def test_readiness(port: int, results: Results) -> None:
    status = curl_status(f"http://localhost:{port}/health/readiness")
    results.record("GET /health/readiness responds", ok=status in (HTTP_OK, HTTP_SERVICE_UNAVAILABLE))


def test_health(port: int, results: Results) -> None:
    r = curl(f"http://localhost:{port}/health")
    ok = False
    if r.returncode == 0:
        try:
            body = json.loads(r.stdout)
            ok = "status" in body and "models" in body
        except json.JSONDecodeError:
            pass
    results.record("GET /health returns JSON with status + models", ok=ok)


def test_openapi(port: int, results: Results) -> None:
    r = curl(f"http://localhost:{port}/openapi.json")
    ok = False
    if r.returncode == 0:
        try:
            spec = json.loads(r.stdout)
            ok = "paths" in spec and "/v1/chat/completions" in spec.get("paths", {})
        except json.JSONDecodeError:
            pass
    results.record("GET /openapi.json has /v1/chat/completions", ok=ok)


def test_openapi_info(port: int, results: Results) -> None:
    r = curl(f"http://localhost:{port}/openapi.json")
    ok = False
    if r.returncode == 0:
        try:
            spec = json.loads(r.stdout)
            info = spec.get("info", {})
            ok = "liter-llm" in info.get("title", "").lower()
        except json.JSONDecodeError:
            pass
    results.record("OpenAPI info.title contains liter-llm", ok=ok)


def test_auth_missing(port: int, results: Results) -> None:
    status = curl_status(f"http://localhost:{port}/v1/models", method="GET")
    results.record("GET /v1/models without auth returns 401", ok=status == HTTP_UNAUTHORIZED)


def test_auth_invalid(port: int, results: Results) -> None:
    status = curl_status(
        f"http://localhost:{port}/v1/models",
        method="GET",
        headers={"Authorization": "Bearer sk-wrong-key"},
    )
    results.record("GET /v1/models with bad key returns 401", ok=status == HTTP_UNAUTHORIZED)


def test_auth_valid(port: int, master_key: str, results: Results) -> None:
    status = curl_status(
        f"http://localhost:{port}/v1/models",
        method="GET",
        headers={"Authorization": f"Bearer {master_key}"},
    )
    results.record("GET /v1/models with master key returns 200", ok=status == HTTP_OK)


def test_nonroot(image: str, results: Results) -> None:
    r = run(["docker", "inspect", "--format", "{{.Config.User}}", image], timeout=10)
    user = r.stdout.strip()
    ok = r.returncode == 0 and user not in {"", "0", "root"}
    results.record("Container runs as non-root", ok=ok)
    if not ok:
        print(f"    User: {user!r}")


def main() -> int:
    parser = argparse.ArgumentParser(description="Docker integration tests for liter-llm")
    parser.add_argument("--image", default="liter-llm:test", help="Docker image to test")
    parser.add_argument("--port", type=int, default=14000, help="Host port for container")
    parser.add_argument("--results-file", default="/tmp/liter-llm-docker-test-results.json")
    args = parser.parse_args()

    image = args.image
    port = args.port
    master_key = "sk-docker-test"
    results = Results()

    print(f"\n{CYAN}liter-llm Docker test suite{RESET}")
    print(f"Image: {image}")
    print(f"Port:  {port}\n")

    print(f"{CYAN}CLI tests:{RESET}")
    test_version(image, results)
    test_help(image, results)
    test_nonroot(image, results)

    print(f"\n{CYAN}Starting container...{RESET}")
    cid = start_container(image, port, env={"LITER_LLM_MASTER_KEY": master_key})
    print(f"  Container: {cid[:12]}")

    if not wait_healthy(port):
        print(f"  {RED}Container failed to start!{RESET}")
        run(["docker", "logs", cid])
        stop_container(cid)
        return 1

    print(f"  {GREEN}Container healthy{RESET}\n")

    try:
        print(f"{CYAN}Health tests:{RESET}")
        test_liveness(port, results)
        test_readiness(port, results)
        test_health(port, results)

        print(f"\n{CYAN}OpenAPI tests:{RESET}")
        test_openapi(port, results)
        test_openapi_info(port, results)

        print(f"\n{CYAN}Auth tests:{RESET}")
        test_auth_missing(port, results)
        test_auth_invalid(port, results)
        test_auth_valid(port, master_key, results)
    finally:
        print(f"\n{CYAN}Stopping container...{RESET}")
        stop_container(cid)

    summary = results.summary()
    summary["image"] = image

    print(f"\n{CYAN}Results:{RESET}")
    print(f"  Total:   {summary['total_tests']}")
    print(f"  Passed:  {GREEN}{summary['passed']}{RESET}")
    color = RED if summary["failed"] else GREEN
    print(f"  Failed:  {color}{summary['failed']}{RESET}")
    print(f"  Rate:    {summary['success_rate']}%")

    if summary["failed_tests"]:
        print(f"\n{RED}Failed tests:{RESET}")
        for t in summary["failed_tests"]:
            print(f"  - {t}")

    with open(args.results_file, "w") as f:
        json.dump(summary, f, indent=2)
    print(f"\nResults written to {args.results_file}")

    return 1 if summary["failed"] > 0 else 0


if __name__ == "__main__":
    sys.exit(main())
