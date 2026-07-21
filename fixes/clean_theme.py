import subprocess

WP = ["wp", "--allow-root", "--path=/var/www/aimoneymachine"]

subprocess.run(
    WP + ["theme", "install", "twentytwentyfive", "--force"],
    capture_output=True,
    text=True,
    timeout=30,
)

with open(
    "/var/www/aimoneymachine/wp-content/themes/twentytwentyfive/functions.php"
) as f:
    php = f.read()

marker = "// AI Money Machine"
if marker in php:
    php = php[: php.index(marker)]

with open(
    "/var/www/aimoneymachine/wp-content/themes/twentytwentyfive/functions.php", "w"
) as f:
    f.write(php)

subprocess.run(WP + ["cache", "flush"], capture_output=True, text=True, timeout=30)
print("Cleaned functions.php")
