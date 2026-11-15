import subprocess

with open(
    "/var/www/aimoneymachine/wp-content/themes/twentytwentyfive/functions.php"
) as f:
    php = f.read()

# Add custom code at the end
php += """
// AI Money Machine - Opulence
add_action('wp_enqueue_scripts', function() {
    wp_add_inline_style('twentytwentyfive-style', '
body,html,.wp-site-blocks,.site,#page,#content,.site-content,.wp-block-group,.wp-block-columns,.wp-block-column,.wp-block-cover,article,section,main,.entry-content,.entry-header,.entry-footer,.widget-area,.sidebar,footer,header,nav,aside{background-color:transparent!important;background-image:none!important}
body{font-family:Segoe UI,-apple-system,BlinkMacSystemFont,Roboto,sans-serif!important;font-size:18px!important;line-height:1.9!important;color:#E8DFD0!important;background:#060608!important}
h1,h2,h3,h4,h5,h6{background:linear-gradient(135deg,#FFD700,#FFC107 30%,#D4AF37 60%,#B8860B 80%,#D4AF37)!important;-webkit-background-clip:text!important;-webkit-text-fill-color:transparent!important;background-clip:text!important;font-weight:800!important;background-color:transparent!important}
h1{font-size:3.2em!important}h2{font-size:2.3em!important;border-bottom:2px solid rgba(212,175,55,0.25)!important;padding-bottom:0.4em!important}h3{font-size:1.6em!important}
p,li,td,th,div,label,figcaption,cite,small,time,dd,dt,dl{color:#CFC4AD!important;background-color:transparent!important}
strong,b,em{background:linear-gradient(135deg,#FFD700,#D4AF37)!important;-webkit-background-clip:text!important;-webkit-text-fill-color:transparent!important;background-clip:text!important;font-weight:800!important;background-color:transparent!important}
a,a:link,a:visited,a:active{color:#C9A84C!important;text-decoration:none!important}a:hover{color:#FFD700!important}
.entry-content{max-width:720px!important;margin:0 auto!important;padding:3em 1.5em!important}
.entry-content blockquote{border-left:3px solid #B8860B!important;padding:1em 1.5em!important;background:rgba(184,134,11,0.04)!important;background-color:rgba(184,134,11,0.04)!important;border-radius:0 10px 10px 0!important;color:#B0A794!important}
.entry-content blockquote *{color:#B0A794!important}
.entry-content pre{background:#0D1017!important;background-color:#0D1017!important;color:#D4C9B0!important;padding:1.5em!important;border-radius:10px!important;border:1px solid rgba(184,134,11,0.15)!important}
.entry-content code{background:rgba(184,134,11,0.08)!important;background-color:rgba(184,134,11,0.08)!important;color:#D4AF37!important;padding:0.15em 0.5em!important;border-radius:5px!important}
.entry-content pre code{background:transparent!important;background-color:transparent!important;color:#D4C9B0!important;border:none!important}
img{max-width:100%!important;height:auto!important;border-radius:10px!important;box-shadow:0 8px 24px rgba(0,0,0,0.5)!important}
th{background:#141820!important;background-color:#141820!important;color:#D4AF37!important;padding:1em 1.2em!important;font-weight:700!important;border-bottom:2px solid rgba(184,134,11,0.3)!important}
td{padding:1em 1.2em!important;border-bottom:1px solid rgba(184,134,11,0.08)!important;color:#BFB7A5!important}
tr:nth-child(even){background:rgba(184,134,11,0.02)!important;background-color:rgba(184,134,11,0.02)!important}
.wp-block-post{background:linear-gradient(160deg,#111520,#0C0F18)!important;background-color:#111520!important;border:1px solid rgba(184,134,11,0.12)!important;border-radius:16px!important;padding:28px!important;margin-bottom:1.5em!important;box-shadow:0 4px 20px rgba(0,0,0,0.3)!important}
.wp-block-post:hover{border-color:rgba(212,175,55,0.35)!important;box-shadow:0 8px 32px rgba(0,0,0,0.4),0 0 40px rgba(184,134,11,0.06)!important}
.wp-block-post-title a{color:#FFF!important;font-weight:700!important}.wp-block-post-title a:hover{color:#FFD700!important}
.wp-block-button__link{background:linear-gradient(135deg,#D4AF37,#B8860B)!important;background-color:#D4AF37!important;color:#0A0A0A!important;border-radius:8px!important;font-weight:800!important}
.wp-block-separator{border-color:rgba(184,134,11,0.2)!important}
header.wp-block-template-part{background:#080A0F!important;background-color:#080A0F!important;border-bottom:1px solid rgba(184,134,11,0.15)!important}
.wp-block-site-title a{background:linear-gradient(135deg,#FFD700,#D4AF37)!important;-webkit-background-clip:text!important;-webkit-text-fill-color:transparent!important;background-clip:text!important;font-weight:900!important;font-size:1.3em!important;background-color:transparent!important}
.wp-block-navigation a{color:#9A9283!important;font-weight:600!important;text-transform:uppercase!important;font-size:0.82em!important;letter-spacing:0.1em!important}.wp-block-navigation a:hover{color:#D4AF37!important}
footer.wp-block-template-part{background:#060709!important;background-color:#060709!important;border-top:1px solid rgba(184,134,11,0.15)!important}
footer.wp-block-template-part *{color:#6B6358!important}footer.wp-block-template-part a{color:#B8860B!important}footer.wp-block-template-part a:hover{color:#D4AF37!important}
footer div[style*=fefcf5]{background:#060709!important;background-color:#060709!important;border-top:1px solid rgba(184,134,11,0.1)!important}
footer div[style*=fefcf5] *{color:#4A453D!important}footer div[style*=fefcf5] a{color:#8B7355!important}footer div[style*=fefcf5] a:hover{color:#B8860B!important}
.widget-area,.sidebar{background:#0D1017!important;background-color:#0D1017!important;border:1px solid rgba(184,134,11,0.1)!important;border-radius:12px!important}
.widget-title{color:#D4AF37!important}.widget-area a{color:#B8860B!important}.widget-area a:hover{color:#D4AF37!important}
.has-white-background-color,.has-base-background-color{background-color:#111520!important}
.has-white-color,.has-base-color{color:#CFC4AD!important}
body{background-color:#060608!important;color:#E8DFD0!important}
a:where(:not(.wp-element-button)){color:#C9A84C!important}a:where(:not(.wp-element-button)):hover{color:#FFD700!important}
blockquote,.wp-block-quote,.wp-block-pullquote{border-left:3px solid #B8860B!important;background:rgba(184,134,11,0.04)!important;background-color:rgba(184,134,11,0.04)!important}
blockquote *,.wp-block-quote *,.wp-block-pullquote *{color:#B0A794!important}
.wp-block-categories a,.wp-block-archives a,.wp-block-latest-posts a,.wp-block-tag-cloud a{color:#B8860B!important}
.wp-block-categories a:hover,.wp-block-archives a:hover,.wp-block-latest-posts a:hover,.wp-block-tag-cloud a:hover{color:#D4AF37!important}
@media(max-width:768px){.entry-content{padding:1.2em!important}h1{font-size:2.2em!important}h2{font-size:1.6em!important}.wp-block-post{padding:20px!important}}
');
}, 99);

// Keyframes
add_action('wp_head', function() {
    echo '<style>
    @keyframes scrollBanner{0%{transform:translateX(0)}100%{transform:translateX(-50%)}}
    @keyframes shimmerSlide{0%{transform:translateX(-100%)}100%{transform:translateX(200%)}}
    @keyframes pulseWord{0%{opacity:0.85;text-shadow:0 0 10px rgba(255,215,0,0.4)}100%{opacity:1;text-shadow:0 0 20px rgba(255,215,0,0.8),0 0 40px rgba(255,215,0,0.3)}}
    @keyframes pulseDot{0%,100%{opacity:0.3;transform:scale(0.8)}50%{opacity:1;transform:scale(1.5)}}
    @keyframes sparkle{0%,100%{opacity:0;transform:scale(0) rotate(0deg)}50%{opacity:1;transform:scale(1) rotate(180deg)}}
    @keyframes flashGold{0%,100%{box-shadow:0 0 0 rgba(212,175,55,0)}50%{box-shadow:0 0 30px rgba(212,175,55,0.3),0 0 60px rgba(212,175,55,0.1)}}
    </style>';
}, 1);

// Top banner with emojis
add_action('wp_body_open', function() {
    $emojis = ['💰','💎','👑','🏆','✨','💫','🌟','⭐','🪙','💍','🔱','⚜️'];
    $words = ['EXCLUSIVE','LUXURY','PREMIUM','ELITE','FORTUNE','EXCELLENCE','DIAMOND','SOVEREIGN','WEALTH','OPULENCE','MAJESTY','GRANDEUR'];
    echo '<div style="background:linear-gradient(90deg,#07090D,#0E1218,#07090D);border-bottom:2px solid #B8860B;padding:1em 0;overflow:hidden;position:relative;width:100%;margin:0;box-shadow:0 4px 30px rgba(184,134,11,0.15);z-index:9999">';
    echo '<div style="position:absolute;top:0;left:0;right:0;bottom:0;background:linear-gradient(90deg,transparent,rgba(255,215,0,0.12),transparent);animation:shimmerSlide 3s ease-in-out infinite;pointer-events:none;z-index:0"></div>';
    echo '<div style="display:flex;animation:scrollBanner 40s linear infinite;white-space:nowrap;position:relative;z-index:1;width:max-content">';
    for ($i = 0; $i < 5; $i++) {
        $idx = 0;
        foreach ($words as $w) {
            $emoji = $emojis[$idx % count($emojis)];
            $delay = $idx * 0.2;
            $dotDelay = $idx * 0.15;
            echo '<span style="color:#FFD700;font-size:1.2em;font-weight:800;letter-spacing:0.3em;text-transform:uppercase;padding:0 0.6em;text-shadow:0 0 12px rgba(255,215,0,0.5);display:inline-block;animation:pulseWord 2.5s ease-in-out infinite alternate;animation-delay:'.$delay.'s">'.$emoji.' '.$w.'</span>';
            echo '<span style="color:#B8860B;padding:0 0.2em;font-size:0.5em;display:inline-block;animation:pulseDot 1.5s ease-in-out infinite;animation-delay:'.$dotDelay.'s;text-shadow:0 0 6px rgba(184,134,11,0.6)">&#9670;</span>';
            $idx++;
        }
    }
    echo '</div></div>';
});
"""

with open(
    "/var/www/aimoneymachine/wp-content/themes/twentytwentyfive/functions.php", "w"
) as f:
    f.write(php)

# Verify syntax
r = subprocess.run(
    [
        "php",
        "-l",
        "/var/www/aimoneymachine/wp-content/themes/twentytwentyfive/functions.php",
    ],
    capture_output=True,
    text=True,
    timeout=10,
)
print(r.stdout.strip())
print(r.stderr.strip())

WP = ["wp", "--allow-root", "--path=/var/www/aimoneymachine"]
subprocess.run(WP + ["cache", "flush"], capture_output=True, text=True, timeout=30)
print("Done")
