#!/usr/bin/env python3
"""
AI Hustle Machine — Proper Per-Symbol Trading Bot
Replaces the buggy Go trading module that shared indicator data across symbols.
"""
import json, os, sys, time, hmac, hashlib, urllib.request, urllib.error, base64
from datetime import datetime, timezone

sys.stdout.reconfigure(encoding="utf-8", errors="replace")

# ── Config ──
SYMBOLS = ["BTCUSDT", "ETHUSDT", "SOLUSDT", "DOGEUSDT", "XRPUSDT", "ADAUSDT", "LINKUSDT"]
QUOTE_ASSET = "USDT"
RSI_PERIOD = 14
SMA_PERIOD = 5
BB_PERIOD = 20
BB_STD = 2
MACD_FAST = 12
MACD_SLOW = 26
MACD_SIGNAL = 9
BUY_RSI = 30
SELL_RSI = 70
STOP_LOSS_PCT = 0.05
TRAILING_STOP_PCT = 0.08
CANDLE_INTERVAL = "5m"
CANDLE_LIMIT = 100

# Load .env
ENV_FILE = os.path.join(os.path.dirname(os.path.abspath(__file__)), '..', '.env')
if os.path.exists(ENV_FILE):
    with open(ENV_FILE) as f:
        for line in f:
            line = line.strip()
            if line and not line.startswith('#') and '=' in line:
                k, v = line.split('=', 1)
                os.environ.setdefault(k.strip(), v.strip())

API_URL = os.environ.get("BINANCE_API_URL", "https://api.binance.us")
API_KEY = os.environ.get("BINANCE_API_KEY", "")
SECRET_KEY = os.environ.get("BINANCE_SECRET_KEY", "")
POSITIONS_FILE = "/opt/aimoneymachine/trading_positions.json"

def log(m):
    t = datetime.now(timezone.utc).strftime("%H:%M:%S")
    print(f"[{t}] {m}", flush=True)

# ── Binance API ──
def binance_get(path, signed=False, params=None):
    url = f"{API_URL}{path}"
    if params:
        qs = "&".join(f"{k}={v}" for k, v in sorted(params.items()))
        url += "?" + qs
    if signed:
        params = params or {}
        params["timestamp"] = int(time.time() * 1000)
        qs = "&".join(f"{k}={v}" for k, v in sorted(params.items()))
        sig = hmac.new(SECRET_KEY.encode(), qs.encode(), hashlib.sha256).hexdigest()
        url = f"{API_URL}{path}?{qs}&signature={sig}"
    req = urllib.request.Request(url, headers={"X-MBX-APIKEY": API_KEY} if signed else {})
    try:
        resp = urllib.request.urlopen(req, timeout=15)
        return json.loads(resp.read())
    except urllib.error.HTTPError as e:
        log(f"  Binance API error {e.code}: {e.read().decode()[:200]}")
        return None

def binance_post(path, params):
    params["timestamp"] = int(time.time() * 1000)
    qs = "&".join(f"{k}={v}" for k, v in sorted(params.items()))
    sig = hmac.new(SECRET_KEY.encode(), qs.encode(), hashlib.sha256).hexdigest()
    url = f"{API_URL}{path}?{qs}&signature={sig}"
    req = urllib.request.Request(url, data=b"", headers={"X-MBX-APIKEY": API_KEY}, method="POST")
    try:
        resp = urllib.request.urlopen(req, timeout=15)
        return json.loads(resp.read())
    except urllib.error.HTTPError as e:
        log(f"  Binance POST error {e.code}: {e.read().decode()[:300]}")
        return None

# ── Indicator Calculations (per symbol!) ──
def calc_rsi(prices, period=RSI_PERIOD):
    if len(prices) < period + 1:
        return 50.0
    gains = losses = 0
    for i in range(-period, 0):
        diff = prices[i] - prices[i - 1]
        if diff >= 0:
            gains += diff
        else:
            losses -= diff
    avg_gain = gains / period
    avg_loss = losses / period
    if avg_loss == 0:
        return 100.0
    rs = avg_gain / avg_loss
    return 100.0 - (100.0 / (1.0 + rs))

def calc_sma(prices, period=SMA_PERIOD):
    if len(prices) < period:
        return sum(prices) / len(prices)
    return sum(prices[-period:]) / period

def calc_bb(prices, period=BB_PERIOD, std=BB_STD):
    if len(prices) < period:
        mid = sum(prices) / len(prices)
        return mid, mid, mid
    mid = sum(prices[-period:]) / period
    variance = sum((p - mid) ** 2 for p in prices[-period:]) / period
    sd = variance ** 0.5
    return mid + std * sd, mid, mid - std * sd

def calc_macd(prices, fast=MACD_FAST, slow=MACD_SLOW, signal=MACD_SIGNAL):
    if len(prices) < slow + signal:
        return 0, 0, 0
    def ema(data, period):
        k = 2 / (period + 1)
        result = [data[0]]
        for i in range(1, len(data)):
            result.append(data[i] * k + result[-1] * (1 - k))
        return result
    ema_fast = ema(prices, fast)
    ema_slow = ema(prices, slow)
    macd_line = [f - s for f, s in zip(ema_fast, ema_slow)]
    signal_line = ema(macd_line, signal)
    hist = macd_line[-1] - signal_line[-1] if len(macd_line) > 0 and len(signal_line) > 0 else 0
    return macd_line[-1] if macd_line else 0, signal_line[-1] if signal_line else 0, hist

# ── Positions ──
def load_positions():
    if os.path.exists(POSITIONS_FILE):
        try:
            with open(POSITIONS_FILE) as f:
                return json.load(f)
        except: pass
    return {}

def save_positions(positions):
    with open(POSITIONS_FILE, "w") as f:
        json.dump(positions, f, indent=2)

# ── Trading Logic ──
def trade_symbol(symbol, positions):
    """Analyze one symbol and decide BUY/SELL/HOLD."""
    # Fetch klines
    klines = binance_get(f"/api/v3/klines", params={
        "symbol": symbol, "interval": CANDLE_INTERVAL, "limit": CANDLE_LIMIT
    })
    if not klines or len(klines) < 30:
        log(f"  {symbol}: Not enough data ({len(klines) if klines else 0} candles)")
        return

    closes = [float(k[4]) for k in klines]
    current_price = closes[-1]

    # Calculate indicators (per symbol!)
    rsi = calc_rsi(closes)
    sma = calc_sma(closes)
    bb_u, bb_m, bb_l = calc_bb(closes)
    macd_line, signal_line, macd_hist = calc_macd(closes)

    log(f"  {symbol}: ${current_price:.4f} | RSI: {rsi:.1f} | SMA({SMA_PERIOD}): ${sma:.2f} | BB: ${bb_l:.2f}-${bb_u:.2f}")

    # Check for open position
    position = positions.get(symbol)
    if position:
        entry = position["entry_price"]
        # Stop-loss check
        if current_price <= entry * (1 - STOP_LOSS_PCT):
            log(f"  ⛔ {symbol}: STOP-LOSS at ${current_price:.4f} (entry ${entry:.4f})")
            if sell(symbol, current_price):
                del positions[symbol]
                save_positions(positions)
            return
        # Trailing stop check
        high_since_entry = position.get("high_since_entry", entry)
        if current_price > high_since_entry:
            position["high_since_entry"] = current_price
        trailing_stop = high_since_entry * (1 - TRAILING_STOP_PCT)
        if current_price <= trailing_stop:
            log(f"  ⛔ {symbol}: TRAILING STOP at ${current_price:.4f} (high ${high_since_entry:.4f})")
            if sell(symbol, current_price):
                del positions[symbol]
                save_positions(positions)
            return
        # SELL signal
        if rsi > SELL_RSI and current_price < sma:
            log(f"  🟢 {symbol}: SELL signal! RSI {rsi:.1f} > {SELL_RSI} & price < SMA")
            if sell(symbol, current_price):
                del positions[symbol]
                save_positions(positions)
            return
        log(f"  📌 {symbol}: HOLDING (entry ${entry:.4f}, stop @ ${max(entry*(1-STOP_LOSS_PCT), trailing_stop):.4f})")
    else:
        # BUY signal
        if rsi < BUY_RSI and current_price > sma:
            log(f"  🟢 {symbol}: BUY signal! RSI {rsi:.1f} < {BUY_RSI} & price > SMA")
            if buy(symbol, current_price):
                positions[symbol] = {
                    "entry_price": current_price,
                    "high_since_entry": current_price,
                    "timestamp": time.time()
                }
                save_positions(positions)
            return
        log(f"  ⏸️  {symbol}: HOLD (RSI {rsi:.1f})")

def buy(symbol, price):
    """Place a market buy order using available USDT balance."""
    balance = binance_get("/api/v3/account", signed=True)
    if not balance:
        log(f"  ❌ Could not fetch balance")
        return
    usdt_bal = 0.0
    for b in balance.get("balances", []):
        if b["asset"] == "USDT":
            usdt_bal = float(b["free"])
            break
    if usdt_bal < 1:
        log(f"  ❌ Insufficient USDT (${usdt_bal:.2f})")
        return False
    spend = usdt_bal * 0.95
    bm = os.environ.get('BINANCE_MIN_NOTIONAL', '5')
    if spend < float(bm):
        log(f"  ⚠️  Spend ${spend:.2f} below min notional, skipping")
        return False
    result = binance_post("/api/v3/order", {
        "symbol": symbol, "side": "BUY", "type": "MARKET", "quoteOrderQty": round(spend, 2)
    })
    if result and result.get("status") == "FILLED":
        filled_qty = float(result.get("executedQty", 0))
        log(f"  ✅ BOUGHT {filled_qty} {symbol.replace('USDT','')} for ${spend:.2f}")
        return True
    elif result and result.get("code"):
        log(f"  ❌ BUY failed: {result.get('msg', str(result))}")
    else:
        log(f"  ❌ BUY failed: {result}")
    return False

def sell(symbol, price):
    """Place a market sell order for the symbol. Returns True if successful."""
    # Get account balance to find the asset qty
    balance = binance_get("/api/v3/account", signed=True)
    if not balance:
        log(f"  ❌ Could not fetch balance")
        return False
    asset = symbol.replace("USDT", "")
    free_qty = 0.0
    for b in balance.get("balances", []):
        if b["asset"] == asset:
            free_qty = float(b["free"])
            break
    if free_qty <= 0:
        log(f"  ⚠️  No {asset} balance to sell ({free_qty})")
        return False
    # Get LOT_SIZE filter for precision
    qty = free_qty
    try:
        info = binance_get("/api/v3/exchangeInfo", params={"symbol": symbol})
        if info and "symbols" in info and len(info["symbols"]) > 0:
            for f in info["symbols"][0].get("filters", []):
                if f["filterType"] == "LOT_SIZE":
                    step = float(f["stepSize"])
                    precision = len(f["stepSize"].rstrip("0").split(".")[-1]) if "." in f["stepSize"] else 0
                    qty = int(free_qty / step) * step
                    qty = round(qty, precision + 1)
                    break
    except:
        qty = round(free_qty, 6)
    if qty <= 0:
        log(f"  ❌ Quantity too small: {qty}")
        return False
    result = binance_post("/api/v3/order", {
        "symbol": symbol, "side": "SELL", "type": "MARKET", "quantity": qty
    })
    if result and result.get("status") == "FILLED":
        fill_qty = float(result.get("executedQty", qty))
        fill_price = float(result.get("cummulativeQuoteQty", 0)) / fill_qty if fill_qty else price
        log(f"  ✅ SOLD {fill_qty} {asset} @ ${fill_price:.4f}")
        return True
    elif result and result.get("code"):
        log(f"  ❌ SELL failed: {result.get('msg', str(result))}")
    else:
        log(f"  ❌ SELL failed: {result}")
    return False

def fetch_sentiment():
    """Fetch Fear & Greed index."""
    try:
        resp = urllib.request.urlopen("https://api.alternative.me/fng/?limit=1", timeout=10)
        data = json.loads(resp.read())
        fng = int(data["data"][0]["value"])
        if fng < 25:
            return fng, "EXTREME FEAR"
        elif fng < 45:
            return fng, "FEAR"
        elif fng < 55:
            return fng, "NEUTRAL"
        elif fng < 75:
            return fng, "GREED"
        else:
            return fng, "EXTREME GREED"
    except:
        return None, "UNKNOWN"

def main():
    log("=" * 55)
    log("  AI HUSTLE MACHINE — Per-Symbol Trading Bot")
    log("=" * 55)

    if not API_KEY or not SECRET_KEY:
        log("  ❌ Missing Binance API credentials")
        return

    # Check exchange connectivity
    exchange = binance_get("/api/v3/exchangeInfo")
    if not exchange:
        log("  ❌ Cannot connect to Binance.US API")
        return
    log(f"  Connected to Binance.US")

    # Fetch sentiment
    fng_val, fng_label = fetch_sentiment()
    if fng_val:
        log(f"  Fear & Greed: {fng_val}/100 ({fng_label})")
    else:
        log(f"  Fear & Greed: UNKNOWN")

    # Fetch balance
    account = binance_get("/api/v3/account", signed=True)
    if account:
        for b in account.get("balances", []):
            bal = float(b.get("free", 0))
            if bal > 0:
                log(f"  Balance: {b['asset']} {bal:.6f}")
    else:
        log(f"  ⚠️  Could not fetch account balance")

    # Trade each symbol
    positions = load_positions()
    for sym in SYMBOLS:
        try:
            trade_symbol(sym, positions)
        except Exception as e:
            log(f"  ❌ Error on {sym}: {e}")
        time.sleep(1)

    # Save positions
    save_positions(positions)

    log(f"{'='*55}")

if __name__ == "__main__":
    while True:
        try:
            main()
            log("Cycle complete. Sleeping 5 minutes...")
            log("")
            time.sleep(300)
        except KeyboardInterrupt:
            break
        except Exception as e:
            log(f"  💥 CRASH: {e}")
            import traceback; traceback.print_exc()
            time.sleep(60)
