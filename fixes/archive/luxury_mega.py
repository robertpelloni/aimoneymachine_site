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

luxury_css = r"""/* MEGA LUXURY THEME - One Trillion Times More Extravagant */
body, html, .wp-site-blocks, .site, #page, #content, .site-content,
.wp-block-group, .wp-block-columns, .wp-block-column, .wp-block-cover,
article, section, main, .entry-content, .entry-header, .entry-footer,
.widget-area, .sidebar, footer, header, nav, aside, div, span, p, li, td, th,
blockquote, pre, code, figure, figcaption, form, input, textarea, select, button,
table, thead, tbody, tfoot, tr, td, th, label, fieldset, legend {
    background-color: transparent !important;
    background-image: none !important;
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
    filter: drop-shadow(0 0 15px rgba(212,175,55,0.4)) !important;
    background-color: transparent !important;
}
h1 { font-size: 3.2em !important; }
h2 { font-size: 2.4em !important; border-bottom: 2px solid rgba(212,175,55,0.3) !important; padding-bottom: 0.5em !important; }
h3 { font-size: 1.8em !important; }
p, li, td, th, span, div, label, figcaption, cite, small, time, dd, dt, dl {
    color: #D8CEBA !important;
}
strong, b, em {
    background: linear-gradient(135deg, #FFD700, #D4AF37) !important;
    -webkit-background-clip: text !important;
    -webkit-text-fill-color: transparent !important;
    background-clip: text !important;
    font-weight: 700 !important;
    background-color: transparent !important;
}
a { color: #7EB8FF !important; text-decoration: none !important; }
a:hover { color: #FFD700 !important; text-shadow: 0 0 20px rgba(255,215,0,0.5) !important; }
.entry-content { max-width: 740px !important; margin: 0 auto !important; padding: 3em 2em !important; }
.entry-content blockquote {
    border-left: 3px solid #D4AF37 !important;
    padding: 1.5em 2em !important;
    background: linear-gradient(135deg, rgba(212,175,55,0.08), rgba(212,175,55,0.02)) !important;
    border-radius: 0 16px 16px 0 !important;
    font-style: italic !important;
    color: #C0B5A0 !important;
    box-shadow: inset 0 0 40px rgba(212,175,55,0.05) !important;
    background-color: rgba(212,175,55,0.05) !important;
}
.entry-content blockquote * { color: #C0B5A0 !important; }
.entry-content pre {
    background: rgba(8,10,16,0.98) !important;
    background-color: rgba(8,10,16,0.98) !important;
    color: #E8DFD0 !important;
    padding: 2em !important;
    border-radius: 16px !important;
    border: 1px solid rgba(212,175,55,0.2) !important;
}
.entry-content code {
    background: rgba(212,175,55,0.12) !important;
    background-color: rgba(212,175,55,0.12) !important;
    color: #FFD700 !important;
    padding: 0.2em 0.6em !important;
    border-radius: 6px !important;
    border: 1px solid rgba(212,175,55,0.2) !important;
}
.entry-content pre code { background: transparent !important; background-color: transparent !important; color: #E8DFD0 !important; border: none !important; }
img { max-width: 100% !important; height: auto !important; border-radius: 16px !important; box-shadow: 0 12px 40px rgba(0,0,0,0.6) !important; }
table { width: 100% !important; border-collapse: collapse !important; border-radius: 16px !important; overflow: hidden !important; }
th {
    background: linear-gradient(135deg, #1A1F2E, #141822) !important;
    background-color: #1A1F2E !important;
    color: #FFD700 !important;
    padding: 1.2em 1.5em !important;
    font-weight: 600 !important;
    border-bottom: 2px solid rgba(212,175,55,0.4) !important;
}
td { padding: 1.2em 1.5em !important; border-bottom: 1px solid rgba(212,175,55,0.1) !important; color: #D0C5B0 !important; }
tr:nth-child(even) { background: rgba(212,175,55,0.03) !important; background-color: rgba(212,175,55,0.03) !important; }
.wp-block-post {
    background: linear-gradient(145deg, rgba(22,27,38,0.98), rgba(10,13,19,0.99)) !important;
    background-color: rgba(22,27,38,0.98) !important;
    border: 1px solid rgba(212,175,55,0.15) !important;
    border-radius: 24px !important;
    padding: 36px !important;
    box-shadow: 0 12px 40px rgba(0,0,0,0.4), 0 0 0 1px rgba(212,175,55,0.08), inset 0 1px 0 rgba(255,255,255,0.02) !important;
}
.wp-block-post:hover {
    border-color: rgba(212,175,55,0.5) !important;
    box-shadow: 0 20px 60px rgba(212,175,55,0.18), 0 0 80px rgba(212,175,55,0.08) !important;
}
.wp-block-post-title a { color: #FFFFFF !important; font-weight: 700 !important; }
.wp-block-post-title a:hover { color: #FFD700 !important; }
.wp-block-button__link {
    background: linear-gradient(135deg, #FFD700, #D4AF37, #B8860B) !important;
    background-color: #D4AF37 !important;
    color: #0A0F1A !important;
    border-radius: 12px !important;
    font-weight: 800 !important;
    box-shadow: 0 6px 20px rgba(212,175,55,0.4) !important;
}
.wp-block-button__link:hover { box-shadow: 0 12px 40px rgba(212,175,55,0.6) !important; }
.wp-block-separator { border-color: rgba(212,175,55,0.25) !important; }
header.wp-block-template-part {
    background: linear-gradient(135deg, #0A0D14, #060810) !important;
    background-color: #0A0D14 !important;
    border-bottom: 1px solid rgba(212,175,55,0.2) !important;
    box-shadow: 0 8px 32px rgba(0,0,0,0.5) !important;
}
.wp-block-site-title a {
    background: linear-gradient(135deg, #FFD700, #D4AF37) !important;
    -webkit-background-clip: text !important;
    -webkit-text-fill-color: transparent !important;
    font-weight: 900 !important;
    font-size: 1.4em !important;
    background-color: transparent !important;
}
.wp-block-navigation a { color: #B0A898 !important; font-weight: 600 !important; text-transform: uppercase !important; font-size: 0.85em !important; }
.wp-block-navigation a:hover { color: #FFD700 !important; }

/* FOOTER - ANIMATED ROTATING BANNER */
footer.wp-block-template-part {
    background: linear-gradient(180deg, #060810, #000) !important;
    background-color: #060810 !important;
    border-top: 1px solid rgba(212,175,55,0.2) !important;
    padding: 0 !important;
    overflow: hidden !important;
}
footer.wp-block-template-part * { color: #FFD700 !important; }
footer.wp-block-template-part a { color: #FFD700 !important; }
footer.wp-block-template-part a:hover { color: #FFFFFF !important; text-shadow: 0 0 20px rgba(255,215,0,0.5) !important; }

/* Animated rotating banner */
.luxury-banner {
    background: linear-gradient(90deg, #0A0F1A, #1A1F2E, #0A0F1A) !important;
    background-color: #0A0F1A !important;
    border-top: 2px solid rgba(212,175,55,0.3) !important;
    border-bottom: 2px solid rgba(212,175,55,0.3) !important;
    padding: 1.5em 0 !important;
    overflow: hidden !important;
    position: relative !important;
}
.luxury-banner::before {
    content: '' !important;
    position: absolute !important;
    top: 0 !important;
    left: -100% !important;
    width: 300% !important;
    height: 100% !important;
    background: linear-gradient(90deg, transparent, rgba(212,175,55,0.1), transparent) !important;
    animation: shimmer 3s infinite !important;
}
@keyframes shimmer {
    0% { transform: translateX(-33%) !important; }
    100% { transform: translateX(33%) !important; }
}
.luxury-banner-text {
    display: flex !important;
    animation: scroll 20s linear infinite !important;
    white-space: nowrap !important;
}
@keyframes scroll {
    0% { transform: translateX(0) !important; }
    100% { transform: translateX(-50%) !important; }
}
.luxury-banner-text span {
    color: #FFD700 !important;
    font-size: 1.2em !important;
    font-weight: 700 !important;
    letter-spacing: 0.3em !important;
    text-transform: uppercase !important;
    padding: 0 2em !important;
    text-shadow: 0 0 20px rgba(212,175,55,0.5) !important;
}
.luxury-banner-text .gold-dot {
    color: #D4AF37 !important;
    padding: 0 0.5em !important;
}

/* Stealth links footer - dark theme */
footer div[style*="fefcf5"] {
    background: #0A0F1A !important;
    background-color: #0A0F1A !important;
    border-top: 1px solid rgba(212,175,55,0.2) !important;
}
footer div[style*="fefcf5"] * {
    color: #6B6358 !important;
}
footer div[style*="fefcf5"] a {
    color: #7EB8FF !important;
}
footer div[style*="fefcf5"] a:hover {
    color: #FFD700 !important;
}

/* Widget area */
.widget-area, .sidebar {
    background: rgba(15,18,25,0.95) !important;
    background-color: rgba(15,18,25,0.95) !important;
    border: 1px solid rgba(212,175,55,0.15) !important;
    border-radius: 20px !important;
}
.widget-title { color: #FFD700 !important; }

/* Color overrides */
.has-white-background-color, .has-base-background-color {
    background-color: rgba(22,27,38,0.98) !important;
}
.has-white-color, .has-base-color {
    color: #E0D5C0 !important;
}
blockquote, .wp-block-quote, .wp-block-pullquote {
    border-left: 3px solid #D4AF37 !important;
    background: rgba(212,175,55,0.05) !important;
    background-color: rgba(212,175,55,0.05) !important;
}
blockquote *, .wp-block-quote *, .wp-block-pullquote * { color: #C0B5A0 !important; }

/* Global style overrides */
body{background-color: #060608 !important; color: #E0D5C0 !important;}
a:where(:not(.wp-element-button)){color: #7EB8FF !important;}
a:where(:not(.wp-element-button)):hover{color: #FFD700 !important;}
h1, h2, h3, h4, h5, h6{background: linear-gradient(135deg, #FFD700, #D4AF37, #B8860B, #D4AF37) !important; -webkit-background-clip: text !important; -webkit-text-fill-color: transparent !important; background-clip: text !important;}
:root :where(.wp-element-button, .wp-block-button__link){background: linear-gradient(135deg, #FFD700, #D4AF37, #B8860B) !important; background-color: #D4AF37 !important; color: #0A0F1A !important;}

@media (max-width: 768px) {
    .entry-content { padding: 1.5em !important; }
    h1 { font-size: 2.2em !important; }
    h2 { font-size: 1.6em !important; }
    .wp-block-post { padding: 24px !important; }
    .luxury-banner-text span { font-size: 1em !important; }
}"""

content += "\n\n// AI Money Machine - MEGA Luxury Theme\n"
content += "add_action('wp_enqueue_scripts', function() {\n"
content += (
    "    wp_add_inline_style('twentytwentyfive-style', '"
    + luxury_css.replace("'", "\\'")
    + "');\n"
)
content += "}, 99);\n"

# Add footer banner via wp_footer hook
content += "\nadd_action('wp_footer', function() {\n"
content += '    echo \'<div class="luxury-banner"><div class="luxury-banner-text">\';\n'
content += "    for ($i = 0; $i < 3; $i++) {\n"
content += "        echo '<span>EXCLUSIVE</span><span class=\"gold-dot\">◆</span>';\n"
content += "        echo '<span>LUXURY</span><span class=\"gold-dot\">◆</span>';\n"
content += "        echo '<span>PREMIUM</span><span class=\"gold-dot\">◆</span>';\n"
content += "        echo '<span>ELITE</span><span class=\"gold-dot\">◆</span>';\n"
content += "        echo '<span>FORTUNE</span><span class=\"gold-dot\">◆</span>';\n"
content += "        echo '<span>EXCELLENCE</span><span class=\"gold-dot\">◆</span>';\n"
content += "    }\n"
content += "    echo '</div></div>';\n"
content += "}, 99);\n"

with open(
    "/var/www/aimoneymachine/wp-content/themes/twentytwentyfive/functions.php", "w"
) as f:
    f.write(content)

subprocess.run(WP + ["cache", "flush"], capture_output=True, text=True, timeout=30)
subprocess.run(
    WP + ["transient", "delete", "--all"], capture_output=True, text=True, timeout=30
)
print("MEGA luxury theme applied with animated banner")
