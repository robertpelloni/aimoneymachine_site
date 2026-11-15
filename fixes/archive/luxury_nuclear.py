import subprocess

WP = ["wp", "--allow-root", "--path=/var/www/aimoneymachine"]

# Refresh theme
subprocess.run(
    WP + ["theme", "install", "twentytwentyfive", "--force"],
    capture_output=True,
    text=True,
    timeout=30,
)

# Read fresh functions.php
with open(
    "/var/www/aimoneymachine/wp-content/themes/twentytwentyfive/functions.php"
) as f:
    content = f.read()

# Nuclear option: override EVERYTHING with !important
luxury_css = r"""/* NUCLEAR OVERRIDE - Every single property with !important */
*, *::before, *::after {
    background-color: transparent !important;
    color: #E0D5C0 !important;
}
body {
    font-family: 'Segoe UI', -apple-system, BlinkMacSystemFont, Roboto, sans-serif !important;
    font-size: 17px !important;
    line-height: 1.85 !important;
    color: #F1E8D0 !important;
    background: linear-gradient(180deg, #060608 0%, #0B0D11 40%, #0E1117 100%) !important;
    background-color: #060608 !important;
}
h1, h2, h3, h4, h5, h6 {
    background: linear-gradient(135deg, #FFD700 0%, #D4AF37 40%, #B8860B 70%, #D4AF37 100%) !important;
    -webkit-background-clip: text !important;
    -webkit-text-fill-color: transparent !important;
    background-clip: text !important;
    font-weight: 700 !important;
    line-height: 1.25 !important;
    filter: drop-shadow(0 0 12px rgba(212,175,55,0.3)) !important;
}
h1 { font-size: 3em !important; margin-bottom: 0.5em !important; }
h2 { font-size: 2.2em !important; padding-bottom: 0.5em !important; border-bottom: 2px solid rgba(212,175,55,0.3) !important; }
h3 { font-size: 1.6em !important; }
p, li, td, th, span, div, label, figcaption, cite, small, time {
    color: #D8CEBA !important;
}
strong, b, em {
    background: linear-gradient(135deg, #FFD700, #D4AF37) !important;
    -webkit-background-clip: text !important;
    -webkit-text-fill-color: transparent !important;
    background-clip: text !important;
    font-weight: 700 !important;
}
a { color: #7EB8FF !important; text-decoration: none !important; }
a:hover { color: #FFD700 !important; text-shadow: 0 0 15px rgba(255,215,0,0.4) !important; }
.entry-content { max-width: 740px !important; margin: 0 auto !important; padding: 3em 2em !important; }
.entry-content blockquote {
    border-left: 3px solid #D4AF37 !important;
    padding: 1.5em 2em !important;
    background: linear-gradient(135deg, rgba(212,175,55,0.08), rgba(212,175,55,0.02)) !important;
    border-radius: 0 16px 16px 0 !important;
    font-style: italic !important;
    color: #C0B5A0 !important;
    box-shadow: inset 0 0 40px rgba(212,175,55,0.05) !important;
}
.entry-content pre {
    background: rgba(8,10,16,0.98) !important;
    color: #E8DFD0 !important;
    padding: 2em !important;
    border-radius: 16px !important;
    border: 1px solid rgba(212,175,55,0.2) !important;
    box-shadow: 0 8px 32px rgba(0,0,0,0.5), inset 0 1px 0 rgba(212,175,55,0.08) !important;
}
.entry-content code {
    background: rgba(212,175,55,0.12) !important;
    padding: 0.2em 0.6em !important;
    border-radius: 6px !important;
    color: #FFD700 !important;
    border: 1px solid rgba(212,175,55,0.2) !important;
}
.entry-content pre code { background: transparent !important; color: #E8DFD0 !important; border: none !important; }
img { max-width: 100% !important; height: auto !important; border-radius: 16px !important; box-shadow: 0 12px 40px rgba(0,0,0,0.6) !important; }
table { width: 100% !important; border-collapse: collapse !important; border-radius: 16px !important; overflow: hidden !important; box-shadow: 0 8px 32px rgba(0,0,0,0.5) !important; }
th {
    background: linear-gradient(135deg, #1A1F2E, #141822) !important;
    color: #FFD700 !important;
    padding: 1.2em 1.5em !important;
    font-weight: 600 !important;
    border-bottom: 2px solid rgba(212,175,55,0.4) !important;
    letter-spacing: 0.05em !important;
    text-transform: uppercase !important;
    font-size: 0.85em !important;
}
td { padding: 1.2em 1.5em !important; border-bottom: 1px solid rgba(212,175,55,0.1) !important; color: #D0C5B0 !important; }
tr:nth-child(even) { background: rgba(212,175,55,0.03) !important; }
tr:hover { background: rgba(212,175,55,0.08) !important; }
.wp-block-post {
    background: linear-gradient(145deg, rgba(22,27,38,0.98), rgba(10,13,19,0.99)) !important;
    border: 1px solid rgba(212,175,55,0.15) !important;
    border-radius: 24px !important;
    padding: 36px !important;
    box-shadow: 0 12px 40px rgba(0,0,0,0.4), 0 0 0 1px rgba(212,175,55,0.08), inset 0 1px 0 rgba(255,255,255,0.02) !important;
    margin-bottom: 2em !important;
}
.wp-block-post:hover {
    border-color: rgba(212,175,55,0.5) !important;
    box-shadow: 0 20px 60px rgba(212,175,55,0.18), 0 0 80px rgba(212,175,55,0.08) !important;
}
.wp-block-post-title a { color: #FFFFFF !important; font-weight: 700 !important; font-size: 1.2em !important; }
.wp-block-post-title a:hover { color: #FFD700 !important; }
.wp-block-button__link {
    background: linear-gradient(135deg, #FFD700, #D4AF37, #B8860B) !important;
    color: #0A0F1A !important;
    border-radius: 12px !important;
    font-weight: 800 !important;
    letter-spacing: 0.05em !important;
    box-shadow: 0 6px 20px rgba(212,175,55,0.4) !important;
    text-transform: uppercase !important;
    font-size: 0.9em !important;
}
.wp-block-button__link:hover { box-shadow: 0 12px 40px rgba(212,175,55,0.6) !important; }
.wp-block-separator { border-color: rgba(212,175,55,0.25) !important; }
.wp-block-cover {
    border-radius: 20px !important;
    overflow: hidden !important;
    box-shadow: 0 16px 48px rgba(0,0,0,0.6) !important;
}
header.wp-block-template-part {
    background: linear-gradient(135deg, #0A0D14, #060810) !important;
    border-bottom: 1px solid rgba(212,175,55,0.2) !important;
    box-shadow: 0 8px 32px rgba(0,0,0,0.5) !important;
    padding: 1em 0 !important;
}
.wp-block-site-title a {
    background: linear-gradient(135deg, #FFD700, #D4AF37) !important;
    -webkit-background-clip: text !important;
    -webkit-text-fill-color: transparent !important;
    font-weight: 900 !important;
    font-size: 1.4em !important;
    letter-spacing: 0.08em !important;
    text-transform: uppercase !important;
}
.wp-block-navigation { padding: 0.5em 0 !important; }
.wp-block-navigation a { color: #B0A898 !important; font-weight: 600 !important; letter-spacing: 0.03em !important; text-transform: uppercase !important; font-size: 0.85em !important; }
.wp-block-navigation a:hover { color: #FFD700 !important; text-shadow: 0 0 12px rgba(255,215,0,0.3) !important; }
footer.wp-block-template-part {
    background: linear-gradient(180deg, #060810, #000) !important;
    border-top: 1px solid rgba(212,175,55,0.2) !important;
    padding: 4em 0 2em !important;
}
footer.wp-block-template-part * { color: #6B6358 !important; }
footer.wp-block-template-part a { color: #7EB8FF !important; }
footer.wp-block-template-part a:hover { color: #FFD700 !important; }
.wp-block-latest-posts__post-title { color: #FFFFFF !important; }
.wp-block-latest-posts__post-title:hover { color: #FFD700 !important; }
.wp-block-categories__count { color: #6B6358 !important; background: rgba(212,175,55,0.1) !important; border-radius: 20px !important; padding: 0.2em 0.8em !important; }
.widget-title { color: #FFD700 !important; }
.widget-area { background: rgba(15,18,25,0.95) !important; border: 1px solid rgba(212,175,55,0.15) !important; border-radius: 20px !important; padding: 2em !important; }
.has-white-background-color, .has-base-background-color, .has-light-gray-background-color, .has-very-light-gray-background-color {
    background-color: rgba(22,27,38,0.98) !important;
}
.has-white-color, .has-base-color, .has-light-gray-color, .has-very-light-gray-color {
    color: #E0D5C0 !important;
}
.wp-block-quote { border-left: 3px solid #D4AF37 !important; background: rgba(212,175,55,0.05) !important; }
.wp-block-quote * { color: #C0B5A0 !important; }
.wp-block-pullquote { border: 2px solid rgba(212,175,55,0.3) !important; background: rgba(212,175,55,0.03) !important; }
.wp-block-pullquote * { color: #C0B5A0 !important; }
.wp-block-media-text { background: rgba(15,18,25,0.95) !important; border-radius: 16px !important; overflow: hidden !important; }
.wp-block-media-text * { color: #D8CEBA !important; }
.wp-block-gallery .wp-block-image { border-radius: 12px !important; overflow: hidden !important; }
figcaption { color: #8A8272 !important; }
.wp-block-archives, .wp-block-categories { background: rgba(15,18,25,0.95) !important; border-radius: 16px !important; padding: 1.5em !important; border: 1px solid rgba(212,175,55,0.1) !important; }
.wp-block-archives a, .wp-block-categories a { color: #7EB8FF !important; }
.wp-block-archives a:hover, .wp-block-categories a:hover { color: #FFD700 !important; }
.wp-block-tag-cloud a { background: rgba(212,175,55,0.1) !important; color: #D4AF37 !important; border-radius: 20px !important; padding: 0.3em 0.8em !important; }
.wp-block-tag-cloud a:hover { background: rgba(212,175,55,0.2) !important; color: #FFD700 !important; }
code { background: rgba(212,175,55,0.1) !important; color: #FFD700 !important; padding: 0.15em 0.4em !important; border-radius: 4px !important; }
pre { background: rgba(8,10,16,0.98) !important; color: #E8DFD0 !important; border: 1px solid rgba(212,175,55,0.15) !important; border-radius: 12px !important; }
@media (max-width: 768px) {
    .entry-content { padding: 1.5em !important; }
    h1 { font-size: 2.2em !important; }
    h2 { font-size: 1.6em !important; }
    .wp-block-post { padding: 24px !important; }
}"""

content += "\n\n// AI Money Machine - Nuclear Luxury Override\n"
content += "add_action('wp_enqueue_scripts', function() {\n"
content += (
    "    wp_add_inline_style('twentytwentyfive-style', '"
    + luxury_css.replace("'", "\\'")
    + "');\n"
)
content += "}, 99);\n"

with open(
    "/var/www/aimoneymachine/wp-content/themes/twentytwentyfive/functions.php", "w"
) as f:
    f.write(content)

subprocess.run(WP + ["cache", "flush"], capture_output=True, text=True, timeout=30)
subprocess.run(
    WP + ["transient", "delete", "--all"], capture_output=True, text=True, timeout=30
)
print("Nuclear luxury override applied")
