import subprocess

with open(
    "/var/www/aimoneymachine/wp-content/themes/twentytwentyfive/functions.php"
) as f:
    php = f.read()

marker = "// AI Money Machine"
if marker in php:
    php = php[: php.index(marker)]

php += r"""
// AI Money Machine - MAXIMUM OPULENCE
add_action('wp_enqueue_scripts', function() {
    wp_add_inline_style('twentytwentyfive-style', '
body,html,.wp-site-blocks,.site,#page,#content,.site-content,.wp-block-group,.wp-block-columns,.wp-block-column,.wp-block-cover,article,section,main,.entry-content,.entry-header,.entry-footer,.widget-area,.sidebar,footer,header,nav,aside{background-color:transparent!important;background-image:none!important}
body{font-family:Georgia,Times New Roman,serif!important;font-size:18px!important;line-height:1.85!important;color:#E8DFD0!important;background:linear-gradient(180deg,#040506 0%,#080A0D 50%,#0A0D12 100%)!important;background-color:#040506!important}
h1,h2,h3,h4,h5,h6{background:linear-gradient(135deg,#FFE55C 0%,#FFD700 15%,#D4AF37 40%,#B8860B 65%,#D4AF37 85%,#FFD700 100%)!important;-webkit-background-clip:text!important;-webkit-text-fill-color:transparent!important;background-clip:text!important;font-weight:800!important;line-height:1.15!important;filter:drop-shadow(0 3px 12px rgba(212,175,55,0.5))!important;background-color:transparent!important;letter-spacing:-0.02em!important}
h1{font-size:3.8em!important;margin-bottom:0.3em!important;text-shadow:0 0 40px rgba(255,215,0,0.3)!important}h2{font-size:2.6em!important;border-bottom:3px solid rgba(212,175,55,0.3)!important;padding-bottom:0.4em!important;margin:1.5em 0 0.5em!important;box-shadow:0 2px 0 rgba(184,134,11,0.15)!important}h3{font-size:1.9em!important;margin:1.2em 0 0.4em!important}
p,li,td,th,div,label,figcaption,cite,small,time,dd,dt,dl{color:#D8CEBA!important;background-color:transparent!important}
strong,b,em{background:linear-gradient(135deg,#FFE55C,#FFD700,#D4AF37)!important;-webkit-background-clip:text!important;-webkit-text-fill-color:transparent!important;background-clip:text!important;font-weight:800!important;background-color:transparent!important;filter:drop-shadow(0 1px 3px rgba(212,175,55,0.3))!important}
a,a:link,a:visited,a:active{color:#D4AF37!important;text-decoration:none!important;border-bottom:1px solid rgba(212,175,55,0.2)!important;transition:all 0.3s ease!important}a:hover{color:#FFD700!important;border-bottom-color:rgba(255,215,0,0.6)!important;text-shadow:0 0 20px rgba(255,215,0,0.4),0 0 40px rgba(255,215,0,0.15)!important}
.entry-content{max-width:720px!important;margin:0 auto!important;padding:3em 1.5em!important}
.entry-content p{margin-bottom:1.5em!important}
.entry-content blockquote{border-left:4px solid #D4AF37!important;padding:1.5em 2em!important;margin:2em 0!important;background:linear-gradient(135deg,rgba(212,175,55,0.08),rgba(212,175,55,0.02))!important;background-color:rgba(212,175,55,0.05)!important;border-radius:0 16px 16px 0!important;color:#C0B5A0!important;font-style:italic!important;box-shadow:inset 0 0 30px rgba(212,175,55,0.03),4px 0 20px rgba(0,0,0,0.2)!important}
.entry-content blockquote *{color:#C0B5A0!important}
.entry-content pre{background:#080B12!important;background-color:#080B12!important;color:#D4C9B0!important;padding:2em!important;border-radius:16px!important;border:1px solid rgba(212,175,55,0.2)!important;box-shadow:inset 0 2px 10px rgba(0,0,0,0.4),0 4px 20px rgba(0,0,0,0.3)!important}
.entry-content code{background:rgba(212,175,55,0.1)!important;background-color:rgba(212,175,55,0.1)!important;color:#FFD700!important;padding:0.2em 0.7em!important;border-radius:8px!important;border:1px solid rgba(212,175,55,0.18)!important;font-weight:600!important}
.entry-content pre code{background:transparent!important;background-color:transparent!important;color:#D4C9B0!important;border:none!important;font-weight:400!important}
img{max-width:100%!important;height:auto!important;border-radius:16px!important;box-shadow:0 16px 48px rgba(0,0,0,0.7),0 0 0 1px rgba(212,175,55,0.1)!important}
table{width:100%!important;border-collapse:collapse!important;border-radius:16px!important;overflow:hidden!important;box-shadow:0 12px 40px rgba(0,0,0,0.5),0 0 0 1px rgba(212,175,55,0.1)!important}
th{background:linear-gradient(180deg,#161B26,#10141E)!important;background-color:#161B26!important;color:#FFD700!important;padding:1.2em 1.5em!important;font-weight:700!important;border-bottom:3px solid rgba(212,175,55,0.4)!important;text-transform:uppercase!important;font-size:0.8em!important;letter-spacing:0.15em!important}
td{padding:1.2em 1.5em!important;border-bottom:1px solid rgba(212,175,55,0.08)!important;color:#C5BBAA!important}
tr:nth-child(even){background:rgba(212,175,55,0.03)!important;background-color:rgba(212,175,55,0.03)!important}
tr:hover{background:rgba(212,175,55,0.07)!important;background-color:rgba(212,175,55,0.07)!important}
.wp-block-post{background:linear-gradient(160deg,#131820 0%,#0D1018 100%)!important;background-color:#131820!important;border:1px solid rgba(212,175,55,0.15)!important;border-radius:24px!important;padding:36px!important;margin-bottom:2em!important;box-shadow:0 12px 40px rgba(0,0,0,0.5),0 0 0 1px rgba(212,175,55,0.05),inset 0 1px 0 rgba(255,255,255,0.03)!important;transition:border-color 0.4s ease,box-shadow 0.4s ease!important}
.wp-block-post:hover{border-color:rgba(212,175,55,0.45)!important;box-shadow:0 20px 60px rgba(0,0,0,0.6),0 0 80px rgba(212,175,55,0.1),0 0 0 1px rgba(212,175,55,0.15)!important}
.wp-block-post-title a{color:#FFFFFF!important;font-weight:700!important;font-size:1.15em!important;letter-spacing:-0.01em!important}.wp-block-post-title a:hover{color:#FFD700!important;text-shadow:0 0 20px rgba(255,215,0,0.3)!important}
.wp-block-button__link{background:linear-gradient(135deg,#FFE55C 0%,#FFD700 25%,#D4AF37 50%,#B8860B 75%,#D4AF37 100%)!important;background-color:#D4AF37!important;color:#0A0A0A!important;border-radius:12px!important;font-weight:900!important;text-transform:uppercase!important;letter-spacing:0.1em!important;font-size:0.85em!important;box-shadow:0 8px 24px rgba(212,175,55,0.4),0 0 0 1px rgba(212,175,55,0.3)!important;padding:16px 36px!important;border:2px solid rgba(255,215,0,0.3)!important}
.wp-block-button__link:hover{box-shadow:0 12px 36px rgba(212,175,55,0.6),0 0 60px rgba(212,175,55,0.2),0 0 0 1px rgba(255,215,0,0.5)!important}
.wp-block-separator{border-color:rgba(212,175,55,0.3)!important;margin:3em 0!important;box-shadow:0 0 20px rgba(212,175,55,0.1)!important}
header.wp-block-template-part{background:linear-gradient(180deg,#0B0E14,#070910)!important;background-color:#0B0E14!important;border-bottom:2px solid rgba(212,175,55,0.2)!important;box-shadow:0 8px 32px rgba(0,0,0,0.6)!important;padding:1em 0!important}
.wp-block-site-title a{background:linear-gradient(135deg,#FFE55C,#FFD700,#D4AF37)!important;-webkit-background-clip:text!important;-webkit-text-fill-color:transparent!important;background-clip:text!important;font-weight:900!important;font-size:1.5em!important;letter-spacing:0.1em!important;text-transform:uppercase!important;background-color:transparent!important;filter:drop-shadow(0 2px 4px rgba(212,175,55,0.3))!important}
.wp-block-navigation a{color:#A09888!important;font-weight:700!important;text-transform:uppercase!important;font-size:0.78em!important;letter-spacing:0.15em!important;border-bottom:none!important;padding:0.5em 1em!important;border-radius:6px!important;transition:all 0.3s ease!important}.wp-block-navigation a:hover{color:#FFD700!important;background:rgba(212,175,55,0.08)!important;text-shadow:0 0 15px rgba(255,215,0,0.3)!important}
footer.wp-block-template-part{background:linear-gradient(180deg,#070910,#030405)!important;background-color:#070910!important;border-top:2px solid rgba(212,175,55,0.2)!important;padding:3em 0 2em!important;box-shadow:0 -8px 32px rgba(0,0,0,0.5)!important}
footer.wp-block-template-part *{color:#8A8272!important}footer.wp-block-template-part a{color:#D4AF37!important;border-bottom:none!important}footer.wp-block-template-part a:hover{color:#FFD700!important;text-shadow:0 0 15px rgba(255,215,0,0.4)!important}
footer div[style*=fefcf5]{background:#070910!important;background-color:#070910!important;border-top:1px solid rgba(212,175,55,0.1)!important}
footer div[style*=fefcf5] *{color:#4A453D!important}footer div[style*=fefcf5] a{color:#6B6358!important;border-bottom:none!important}footer div[style*=fefcf5] a:hover{color:#8B7355!important}
.widget-area,.sidebar{background:linear-gradient(160deg,#131820,#0D1018)!important;background-color:#131820!important;border:1px solid rgba(212,175,55,0.15)!important;border-radius:20px!important;padding:2em!important;box-shadow:0 8px 32px rgba(0,0,0,0.4)!important}
.widget-title{color:#FFD700!important;font-weight:800!important;text-transform:uppercase!important;letter-spacing:0.12em!important;font-size:0.85em!important;border-bottom:2px solid rgba(212,175,55,0.3)!important;padding-bottom:0.8em!important;margin-bottom:1em!important}
.widget-area a{color:#B8860B!important;border-bottom:none!important}.widget-area a:hover{color:#FFD700!important;text-shadow:0 0 10px rgba(255,215,0,0.3)!important}
.wp-block-latest-posts__post-title a{color:#FFF!important;border-bottom:none!important;font-weight:700!important}.wp-block-latest-posts__post-title a:hover{color:#FFD700!important}
.wp-block-cover{border-radius:20px!important;overflow:hidden!important;box-shadow:0 20px 60px rgba(0,0,0,0.7),0 0 0 1px rgba(212,175,55,0.1)!important}
.wp-block-categories a,.wp-block-archives a,.wp-block-latest-posts a,.wp-block-tag-cloud a{color:#B8860B!important;border-bottom:none!important}
.wp-block-categories a:hover,.wp-block-archives a:hover,.wp-block-latest-posts a:hover,.wp-block-tag-cloud a:hover{color:#FFD700!important}
.wp-block-categories__count{background:rgba(212,175,55,0.1)!important;background-color:rgba(212,175,55,0.1)!important;color:#D4AF37!important;border-radius:20px!important;padding:0.2em 0.8em!important;font-size:0.85em!important}
blockquote,.wp-block-quote,.wp-block-pullquote{border-left:4px solid #D4AF37!important;background:linear-gradient(135deg,rgba(212,175,55,0.08),rgba(212,175,55,0.02))!important;background-color:rgba(212,175,55,0.05)!important;border-radius:0 16px 16px 0!important}
blockquote *,.wp-block-quote *,.wp-block-pullquote *{color:#C0B5A0!important}
.has-white-background-color,.has-base-background-color{background-color:#131820!important}
.has-white-color,.has-base-color{color:#D8CEBA!important}
a:where(:not(.wp-element-button)){color:#D4AF37!important;border-bottom:1px solid rgba(212,175,55,0.15)!important}a:where(:not(.wp-element-button)):hover{color:#FFD700!important}
@media(max-width:768px){.entry-content{padding:1.5em!important}h1{font-size:2.5em!important}h2{font-size:1.9em!important}h3{font-size:1.4em!important}.wp-block-post{padding:24px!important;margin-bottom:1.5em!important}.wp-block-button__link{padding:14px 28px!important}}
');
}, 99);

// Keyframes
add_action('wp_head', function() {
    echo '<style>
    @keyframes scrollBanner{0%{transform:translateX(0)}100%{transform:translateX(-50%)}}
    @keyframes shimmerSlide{0%{transform:translateX(-100%)}100%{transform:translateX(200%)}}
    @keyframes pulseWord{0%{opacity:0.8;text-shadow:0 0 8px rgba(255,215,0,0.4)}100%{opacity:1;text-shadow:0 0 25px rgba(255,215,0,1),0 0 50px rgba(255,215,0,0.5),0 0 75px rgba(255,215,0,0.2)}}
    @keyframes pulseDot{0%,100%{opacity:0.2;transform:scale(0.6)}50%{opacity:1;transform:scale(1.8)}}
    </style>';
}, 1);

// Banner function
function ai_money_machine_banner() {
    $emojis = array('💰','💎','👑','🏆','✨','💫','🌟','⭐','🪙','💍','🔱','⚜️');
    $words = array('EXCLUSIVE','LUXURY','PREMIUM','ELITE','FORTUNE','EXCELLENCE','DIAMOND','SOVEREIGN','WEALTH','OPULENCE','MAJESTY','GRANDEUR');
    echo '<div style="background:linear-gradient(90deg,#060810,#0E1420,#060810);border-top:2px solid #B8860B;border-bottom:2px solid #B8860B;padding:1.2em 0;overflow:hidden;position:relative;width:100%;margin:0;box-shadow:0 0 50px rgba(184,134,11,0.2),inset 0 0 30px rgba(184,134,11,0.04)">';
    echo '<div style="position:absolute;top:0;left:0;right:0;bottom:0;background:linear-gradient(90deg,transparent,rgba(255,215,0,0.12),transparent);animation:shimmerSlide 3.5s ease-in-out infinite;pointer-events:none;z-index:0"></div>';
    echo '<div style="display:flex;animation:scrollBanner 45s linear infinite;white-space:nowrap;position:relative;z-index:1;width:max-content">';
    for ($i = 0; $i < 5; $i++) {
        $idx = 0;
        foreach ($words as $w) {
            $emoji = $emojis[$idx % count($emojis)];
            $delay = $idx * 0.18;
            $dotDelay = $idx * 0.12;
            echo '<span style="color:#FFD700;font-size:1.2em;font-weight:900;letter-spacing:0.35em;text-transform:uppercase;padding:0 0.5em;text-shadow:0 0 12px rgba(255,215,0,0.6),0 0 25px rgba(255,215,0,0.25);display:inline-block;animation:pulseWord 3s ease-in-out infinite alternate;animation-delay:'.$delay.'s">'.$emoji.' '.$w.'</span>';
            echo '<span style="color:#B8860B;padding:0 0.15em;font-size:0.45em;display:inline-block;animation:pulseDot 1.8s ease-in-out infinite;animation-delay:'.$dotDelay.'s;text-shadow:0 0 8px rgba(184,134,11,0.8)">&#9670;</span>';
            $idx++;
        }
    }
    echo '</div></div>';
}

add_action('wp_body_open', function() { ai_money_machine_banner(); });
add_action('wp_footer', function() { ai_money_machine_banner(); });
"""

with open(
    "/var/www/aimoneymachine/wp-content/themes/twentytwentyfive/functions.php", "w"
) as f:
    f.write(php)

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

WP = ["wp", "--allow-root", "--path=/var/www/aimoneymachine"]
subprocess.run(WP + ["cache", "flush"], capture_output=True, text=True, timeout=30)
print("Done - MAXIMUM OPULENCE")
