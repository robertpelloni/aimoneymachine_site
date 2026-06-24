import json
import os
import sys

try:
    from atproto import Client
except Exception as e:
    print(f"atproto import failed: {e}")
    sys.exit(1)


def get_credentials():
    try:
        cred_path = "/opt/aimoneymachine/.bluesky_credentials.json"
        if os.path.exists(cred_path):
            data = json.load(open(cred_path))
            return data.get("handle", ""), data.get(
                "app_password", data.get("password", "")
            )
    except Exception as e:
        print(f"Credential read failed: {e}")
    return "", ""


def main():
    handle, app_password = get_credentials()
    if not handle or not app_password:
        print(
            "No Bluesky credentials found. Set BLUESKY_HANDLE and BLUESKY_APP_PASSWORD."
        )
        sys.exit(1)

    text = ""
    queue_path = "/opt/aimoneymachine/dispatcher/social_queue.json"
    if not sys.stdin.isatty():
        text = sys.stdin.read().strip()
    elif os.path.exists(queue_path):
        try:
            queue = json.load(open(queue_path))
            pending = queue.get("pending_tweets", [])
            if pending:
                text = pending[0].get("content", "")
        except Exception:
            pass

    if not text:
        print("No content to post")
        sys.exit(0)

    client = Client()
    try:
        profile = client.login(handle, app_password)
        print(f"Logged in as: {profile.display_name} ({profile.handle})")
    except Exception as e:
        print(f"Login failed: {e}")
        sys.exit(1)

    try:
        post = client.send_post(text=text)
        print(f"Posted: {post.uri}")
    except Exception as e:
        print(f"Post failed: {e}")
        sys.exit(1)


if __name__ == "__main__":
    main()
