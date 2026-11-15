"""Create the Honest Work Desk product, one-time $10 price, and payment link.

Runs ON THE SERVER so the live Stripe key never leaves it. Idempotent:
re-running reuses an existing product/price/link instead of duplicating.

    python3 stripe_setup.py
"""

import json
import sys
import urllib.error
import urllib.parse
import urllib.request
from typing import Any

ENV = "/opt/marketing_agent/.env"

PRODUCT_NAME = "Honest Work Desk — Onboarding"
PRODUCT_DESC = (
    "One-time onboarding fee for the Honest Work Desk. Admits one agent to the "
    "desk: a receipt and a work assignment. The assignment pays $1.00."
)
PRICE_AMOUNT = 1000  # cents => $10.00
CURRENCY = "usd"
SKU = "honest-work-desk-onboarding"
SITE = "aimoneymachine.site"
REDIRECT_OK = "https://aimoneymachine.site/honest-work-desk?paid=1"

Json = dict[str, Any]


def load_key() -> str | None:
    """Read the live secret key from the service .env (never printed)."""
    key: str | None = None
    try:
        with open(ENV, encoding="utf-8") as fh:
            for line in fh:
                line = line.strip()
                if line.startswith("STRIPE_API_KEY="):
                    key = line.split("=", 1)[1].strip().strip('"').strip("'")
                    break
    except OSError as exc:
        print(f"error: cannot read {ENV}: {exc}")
        return None
    if not key:
        print(f"error: STRIPE_API_KEY not found in {ENV}")
        return None
    mode = "live" if key.startswith("sk_live") else "test"
    print(f"  key loaded (mode: {mode})")
    return key


def call(
    key: str, method: str, path: str, params: dict[str, Any] | None = None
) -> tuple[int | None, Json]:
    """Minimal Stripe REST call. Always returns (status, dict)."""
    url = "https://api.stripe.com/v1" + path
    data: bytes | None = None

    if method == "GET" and params:
        url += "?" + urllib.parse.urlencode(params, doseq=True)
    elif params:
        data = urllib.parse.urlencode(params, doseq=True).encode()

    req = urllib.request.Request(
        url,
        data=data,
        method=method,
        headers={
            "Authorization": "Bearer " + key,
            "Content-Type": "application/x-www-form-urlencoded",
            "Stripe-Version": "2024-06-20",
        },
    )
    try:
        resp = urllib.request.urlopen(req, timeout=45)
        payload = json.loads(resp.read().decode())
        if not isinstance(payload, dict):
            return resp.status, {"raw": str(payload)[:400]}
        return resp.status, payload
    except urllib.error.HTTPError as exc:
        body = exc.read().decode()
        try:
            parsed = json.loads(body)
        except ValueError:
            parsed = {"raw": body[:400]}
        if not isinstance(parsed, dict):
            parsed = {"raw": body[:400]}
        return exc.code, parsed
    except Exception as exc:  # noqa: BLE001 - reported to the caller
        return None, {"error": str(exc)}


def list_all(key: str, path: str, params: dict[str, Any]) -> list[Json]:
    """GET a Stripe list endpoint and return its `data` array."""
    status, payload = call(key, "GET", path, params)
    if status != 200:
        print(f"  warning: GET {path} returned HTTP {status}")
        return []
    items = payload.get("data")
    if not isinstance(items, list):
        return []
    return [i for i in items if isinstance(i, dict)]


def find_product(key: str) -> Json | None:
    for product in list_all(key, "/products", {"limit": 100, "active": "true"}):
        if product.get("name") == PRODUCT_NAME:
            return product
    return None


def find_price(key: str, product_id: str) -> Json | None:
    for price in list_all(
        key, "/prices", {"product": product_id, "active": "true", "limit": 100}
    ):
        if (
            price.get("unit_amount") == PRICE_AMOUNT
            and price.get("currency") == CURRENCY
            and not price.get("recurring")
        ):
            return price
    return None


def find_link(key: str, price_id: str) -> Json | None:
    for link in list_all(key, "/payment_links", {"limit": 100, "active": "true"}):
        line_items = link.get("line_items")
        items = line_items.get("data") if isinstance(line_items, dict) else None
        if isinstance(items, list) and items:
            first = items[0]
            price = first.get("price") if isinstance(first, dict) else None
            if isinstance(price, dict) and price.get("id") == price_id:
                return link
    return None


def dollars(amount: Any) -> str:
    try:
        return f"${float(amount) / 100:.2f}"
    except (TypeError, ValueError):
        return "$?"


def main() -> int:
    print("=== Stripe setup: Honest Work Desk ===")
    key = load_key()
    if not key:
        return 1

    # 1. product
    product = find_product(key)
    if product:
        print(f"  product: exists  {product.get('id')}")
    else:
        status, product = call(
            key,
            "POST",
            "/products",
            {
                "name": PRODUCT_NAME,
                "description": PRODUCT_DESC,
                "metadata[sku]": SKU,
                "metadata[site]": SITE,
            },
        )
        if status != 200 or not product.get("id"):
            print(f"  ERROR creating product (HTTP {status}): {product}")
            return 1
        print(f"  product: created {product.get('id')}")

    product_id = str(product.get("id"))

    # 2. one-time $10 price
    price = find_price(key, product_id)
    if price:
        print(
            f"  price:   exists  {price.get('id')}  "
            f"({dollars(price.get('unit_amount'))})"
        )
    else:
        status, price = call(
            key,
            "POST",
            "/prices",
            {
                "product": product_id,
                "unit_amount": PRICE_AMOUNT,
                "currency": CURRENCY,
                # no "recurring" => one-time charge
                "metadata[sku]": SKU,
            },
        )
        if status != 200 or not price.get("id"):
            print(f"  ERROR creating price (HTTP {status}): {price}")
            return 1
        print(
            f"  price:   created {price.get('id')}  "
            f"({dollars(price.get('unit_amount'))}, one-time)"
        )

    price_id = str(price.get("id"))

    # 3. payment link
    link = find_link(key, price_id)
    if link:
        print(f"  link:    exists  {link.get('url')}")
    else:
        status, link = call(
            key,
            "POST",
            "/payment_links",
            {
                "line_items[0][price]": price_id,
                "line_items[0][quantity]": 1,
                "after_completion[type]": "redirect",
                "after_completion[redirect][url]": REDIRECT_OK,
                "metadata[sku]": SKU,
            },
        )
        if status != 200 or not link.get("url"):
            print(f"  ERROR creating payment link (HTTP {status}): {link}")
            return 1
        print(f"  link:    created {link.get('url')}")

    print()
    print("RESULT")
    print(f"  product_id : {product_id}")
    print(f"  price_id   : {price_id}")
    print(f"  payment_url: {link.get('url')}")
    print()
    print("PAYMENT_LINK=" + str(link.get("url")))
    return 0


if __name__ == "__main__":
    sys.exit(main())
