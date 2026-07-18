import subprocess

WP = ["wp", "--allow-root", "--path=/var/www/aimoneymachine"]

# Refresh theme
subprocess.run(
    WP + ["theme", "install", "twentytwentyfive", "--force"],
    capture_output=True,
    text=True,
    timeout=30,
)

with open(
    "/var/www/aimoneymachine/wp-content/themes/twentytwentyfive/functions.php"
) as f:
    content = f.read()

luxury_css = r"""/* GENUINE OPULENCE */
body, html, .wp-site-blocks, .site, #page, #content, .site-content,
.wp-block-group, .wp-block-columns, .wp-block-column, .wp-block-cover,
article, section, main, .entry-content, .entry-header, .entry-footer,
.widget-area, .sidebar, footer, header, nav, aside {
    background-color: transparent !important;
    background-image: none !important;
}
body {
    font-family: 'Segoe UI', -apple-system, BlinkMacSystemFont, Roboto, sans-serif !important;
    font-size: 18px !important;
    line-height: 1.9 !important;
    color: #E8DFD0 !important;
    background: #060608 !important;
}
h1, h2, h3, h4, h5, h6 {
    background: linear-gradient(135deg, #FFD700 0%, #FFC107 30%, #D4AF37 60%, #B8860B 80%, #D4AF37 100%) !important;
    -webkit-background-clip: text !important;
    -webkit-text-fill-color: transparent !important;
    background-clip: text !important;
    font-weight: 800 !important;
    line-height: 1.2 !important;
    filter: drop-shadow(0 2px 4px rgba(0,0,0,0.5)) !important;
    background-color: transparent !important;
    letter-spacing: -0.01em !important;
}
h1 { font-size: 3.2em !important; margin-bottom: 0.4em !important; }
h2 { font-size: 2.3em !important; border-bottom: 2px solid rgba(212,175,55,0.25) !important; padding-bottom: 0.4em !important; margin: 1.8em 0 0.6em !important; }
h3 { font-size: 1.6em !important; margin: 1.4em 0 0.5em !important; }
p, li, td, th, div, span:not(.gd), label, figcaption, cite, small, time, dd, dt, dl {
    color: #CFC4AD !important;
    background-color: transparent !important;
}
strong, b, em {
    background: linear-gradient(135deg, #FFD700, #D4AF37) !important;
    -webkit-background-clip: text !important;
    -webkit-text-fill-color: transparent !important;
    background-clip: text !important;
    font-weight: 800 !important;
    background-color: transparent !important;
}
a, a:link, a:visited, a:active {
    color: #C9A84C !important;
    text-decoration: none !important;
}
a:hover {
    color: #FFD700 !important;
}
.entry-content { max-width: 720px !important; margin: 0 auto !important; padding: 3em 1.5em !important; }
.entry-content p { margin-bottom: 1.4em !important; }
.entry-content blockquote {
    border-left: 3px solid #B8860B !important;
    padding: 1em 1.5em !important;
    margin: 1.8em 0 !important;
    background: rgba(184,134,11,0.04) !important;
    background-color: rgba(184,134,11,0.04) !important;
    border-radius: 0 10px 10px 0 !important;
    color: #B0A794 !important;
}
.entry-content blockquote * { color: #B0A794 !important; }
.entry-content pre {
    background: #0D1017 !important;
    background-color: #0D1017 !important;
    color: #D4C9B0 !important;
    padding: 1.5em !important;
    border-radius: 10px !important;
    border: 1px solid rgba(184,134,11,0.15) !important;
}
.entry-content code {
    background: rgba(184,134,11,0.08) !important;
    background-color: rgba(184,134,11,0.08) !important;
    color: #D4AF37 !important;
    padding: 0.15em 0.5em !important;
    border-radius: 5px !important;
    border: 1px solid rgba(184,134,11,0.12) !important;
}
.entry-content pre code { background: transparent !important; background-color: transparent !important; color: #D4C9B0 !important; border: none !important; }
img { max-width: 100% !important; height: auto !important; border-radius: 10px !important; box-shadow: 0 8px 24px rgba(0,0,0,0.5) !important; }
table { width: 100% !important; border-collapse: collapse !important; border-radius: 10px !important; overflow: hidden !important; }
th {
    background: #141820 !important;
    background-color: #141820 !important;
    color: #D4AF37 !important;
    padding: 1em 1.2em !important;
    font-weight: 700 !important;
    border-bottom: 2px solid rgba(184,134,11,0.3) !important;
    text-transform: uppercase !important;
    font-size: 0.85em !important;
    letter-spacing: 0.08em !important;
}
td { padding: 1em 1.2em !important; border-bottom: 1px solid rgba(184,134,11,0.08) !important; color: #BFB7A5 !important; }
tr:nth-child(even) { background: rgba(184,134,11,0.02) !important; background-color: rgba(184,134,11,0.02) !important; }
.wp-block-post {
    background: linear-gradient(160deg, #111520 0%, #0C0F18 100%) !important;
    background-color: #111520 !important;
    border: 1px solid rgba(184,134,11,0.12) !important;
    border-radius: 16px !important;
    padding: 28px !important;
    margin-bottom: 1.5em !important;
    box-shadow: 0 4px 20px rgba(0,0,0,0.3) !important;
    transition: border-color 0.3s ease, box-shadow 0.3s ease !important;
}
.wp-block-post:hover {
    border-color: rgba(212,175,55,0.35) !important;
    box-shadow: 0 8px 32px rgba(0,0,0,0.4), 0 0 40px rgba(184,134,11,0.06) !important;
}
.wp-block-post-title a { color: #FFFFFF !important; font-weight: 700 !important; }
.wp-block-post-title a:hover { color: #FFD700 !important; }
.wp-block-button__link {
    background: linear-gradient(135deg, #D4AF37, #B8860B) !important;
    background-color: #D4AF37 !important;
    color: #0A0A0A !important;
    border-radius: 8px !important;
    font-weight: 800 !important;
    text-transform: uppercase !important;
    letter-spacing: 0.06em !important;
    font-size: 0.9em !important;
    box-shadow: 0 4px 12px rgba(0,0,0,0.3) !important;
}
.wp-block-button__link:hover { box-shadow: 0 6px 20px rgba(184,134,11,0.3) !important; }
.wp-block-separator { border-color: rgba(184,134,11,0.2) !important; }
header.wp-block-template-part {
    background: #080A0F !important;
    background-color: #080A0F !important;
    border-bottom: 1px solid rgba(184,134,11,0.15) !important;
    box-shadow: 0 2px 12px rgba(0,0,0,0.4) !important;
}
.wp-block-site-title a {
    background: linear-gradient(135deg, #FFD700, #D4AF37) !important;
    -webkit-background-clip: text !important;
    -webkit-text-fill-color: transparent !important;
    background-clip: text !important;
    font-weight: 900 !important;
    font-size: 1.3em !important;
    letter-spacing: 0.06em !important;
    text-transform: uppercase !important;
    background-color: transparent !important;
}
.wp-block-navigation a { color: #9A9283 !important; font-weight: 600 !important; text-transform: uppercase !important; font-size: 0.82em !important; letter-spacing: 0.1em !important; }
.wp-block-navigation a:hover { color: #D4AF37 !important; }

/* FOOTER */
footer.wp-block-template-part {
    background: #060709 !important;
    background-color: #060709 !important;
    border-top: 1px solid rgba(184,134,11,0.15) !important;
    padding: 2.5em 0 2em !important;
}
footer.wp-block-template-part * { color: #6B6358 !important; }
footer.wp-block-template-part a { color: #B8860B !important; }
footer.wp-block-template-part a:hover { color: #D4AF37 !important; }

/* Stealth links */
footer div[style*="fefcf5"] {
    background: #060709 !important;
    background-color: #060709 !important;
    border-top: 1px solid rgba(184,134,11,0.1) !important;
}
footer div[style*="fefcf5"] * { color: #4A453D !important; }
footer div[style*="fefcf5"] a { color: #8B7355 !important; }
footer div[style*="fefcf5"] a:hover { color: #B8860B !important; }

/* Widgets */
.widget-area, .sidebar {
    background: #0D1017 !important;
    background-color: #0D1017 !important;
    border: 1px solid rgba(184,134,11,0.1) !important;
    border-radius: 12px !important;
    padding: 1.5em !important;
}
.widget-title { color: #D4AF37 !important; font-weight: 700 !important; }
.widget-area a { color: #B8860B !important; }
.widget-area a:hover { color: #D4AF37 !important; }

/* WP global overrides */
.has-white-background-color, .has-base-background-color { background-color: #111520 !important; }
.has-white-color, .has-base-color { color: #CFC4AD !important; }
body { background-color: #060608 !important; color: #E8DFD0 !important; }
a:where(:not(.wp-element-button)) { color: #C9A84C !important; }
a:where(:not(.wp-element-button)):hover { color: #FFD700 !important; }
h1,h2,h3,h4,h5,h6 { background: linear-gradient(135deg,#FFD700,#D4AF37,#B8860B,#D4AF37) !important; -webkit-background-clip:text !important; -webkit-text-fill-color:transparent !important; background-clip:text !important; }
:root :where(.wp-element-button,.wp-block-button__link){background:linear-gradient(135deg,#D4AF37,#B8860B)!important;background-color:#D4AF37!important;color:#0A0A0A!important;}
blockquote,.wp-block-quote,.wp-block-pullquote{border-left:3px solid #B8860B!important;background:rgba(184,134,11,0.04)!important;background-color:rgba(184,134,11,0.04)!important;}
blockquote *,.wp-block-quote *,.wp-block-pullquote *{color:#B0A794!important;}
.wp-block-categories a,.wp-block-archives a,.wp-block-latest-posts a,.wp-block-tag-cloud a{color:#B8860B!important;}
.wp-block-categories a:hover,.wp-block-archives a:hover,.wp-block-latest-posts a:hover,.wp-block-tag-cloud a:hover{color:#D4AF37!important;}
@media(max-width:768px){.entry-content{padding:1.2em!important;}h1{font-size:2.2em!important;}h2{font-size:1.6em!important;}.wp-block-post{padding:20px!important;}}"""

content += "\n\n// AI Money Machine - Genuine Opulence\n"
content += "add_action('wp_enqueue_scripts', function() {\n"
content += (
    "    wp_add_inline_style('twentytwentyfive-style', '"
    + luxury_css.replace("'", "\\'")
    + "');\n"
)
content += "}, 99);\n"

# Banner with INLINE styles for guaranteed animation
content += r"""
add_action('wp_footer', function() {
    echo '<div style="background:#080A0F;border-top:2px solid #B8860B;border-bottom:2px solid #B8860B;padding:1.2em 0;overflow:hidden;position:relative;width:100%;margin:0;box-shadow:0 0 40px rgba(184,134,11,0.15),inset 0 0 30px rgba(184,134,11,0.03)">';
    echo '<div style="position:absolute;top:0;left:0;right:0;bottom:0;background:linear-gradient(90deg,transparent,rgba(212,175,55,0.08),transparent);animation:shimmerSlide 4s ease-in-out infinite;"></div>';
    echo '<div style="display:flex;animation:scrollBanner 35s linear infinite;white-space:nowrap;position:relative;z-index:1;width:max-content">';
    $words = array('EXCLUSIVE','LUXURY','PREMIUM','ELITE','FORTUNE','EXCELLENCE','DIAMOND','SOVEREIGN','WEALTH','OPULENCE','MAJESTY','GRANDEUR');
    for ($i = 0; $i < 5; $i++) {
        foreach ($words as $w) {
            echo '<span style="color:#FFD700;font-size:1.3em;font-weight:800;letter-spacing:0.35em;text-transform:uppercase;padding:0 1.2em;text-shadow:0 0 15px rgba(255,215,0,0.5),0 0 30px rgba(255,215,0,0.25);display:inline-block;animation:pulseWord 2.5s ease-in-out infinite alternate;">'.$w.'</span>';
            echo '<span style="color:#B8860B;padding:0 0.6em;font-size:0.6em;display:inline-block;animation:pulseDot 1.8s ease-in-out infinite;text-shadow:0 0 8px rgba(184,134,11,0.6)">&#9670;</span>';
        }
    }
    echo '</div></div>';
});
"""

with open(
    "/var/www/aimoneymachine/wp-content/themes/twentytwentyfive/functions.php", "w"
) as f:
    f.write(content)

# Now add the keyframes animation CSS via a separate style tag in the header
# since wp_head might not get the keyframes
with open(
    "/var/www/aimoneymachine/wp-content/themes/twentytwentyfive/functions.php", "a"
) as f:
    f.write(r"""

// Add keyframes for banner animation
add_action('wp_head', function() {
    echo '<style>
    @keyframes scrollBanner {
        0% { transform: translateX(0); }
        100% { transform: translateX(-50%); }
    }
    @keyframes shimmerSlide {
        0% { transform: translateX(-100%); }
        100% { transform: translateX(100%); }
    }
    @keyframes pulseWord {
        0% { text-shadow: 0 0 15px rgba(255,215,0,0.5), 0 0 30px rgba(255,215,0,0.25); }
        100% { text-shadow: 0 0 25px rgba(255,215,0,0.8), 0 0 50px rgba(255,215,0,0.4), 0 0 70px rgba(255,215,0,0.15); }
    }
    @keyframes pulseDot {
        0%, 100% { opacity: 0.4; transform: scale(1); }
        50% { opacity: 1; transform: scale(1.4); }
    }
    </style>';
}, 1);
""")

subprocess.run(WP + ["cache", "flush"], capture_output=True, text=True, timeout=30)
subprocess.run(
    WP + ["transient", "delete", "--all"], capture_output=True, text=True, timeout=30
)
print("GENUINE OPULENCE applied")
