import subprocess

with open(
    "/var/www/aimoneymachine/wp-content/themes/twentytwentyfive/functions.php"
) as f:
    php = f.read()

marker = "// AI Money Machine"
if marker in php:
    php = php[: php.index(marker)]

php += r"""
// AI Money Machine - INFINITE WEALTH (Compact)
add_action('wp_enqueue_scripts', function() {
    wp_add_inline_style('twentytwentyfive-style', '
*,*::before,*::after{background-color:transparent!important;background-image:none!important}
body{font-family:Palatino Linotype,Book Antiqua,Palatino,Georgia,serif!important;font-size:16px!important;line-height:1.75!important;color:#E0D5C0!important;background:linear-gradient(180deg,#010102 0%,#040608 20%,#080C14 50%,#0C1018 80%,#0E1420 100%)!important;background-color:#010102!important}
h1,h2,h3,h4,h5,h6{background:linear-gradient(135deg,#FFFFFF 0%,#FFFACD 8%,#FFE55C 18%,#FFD700 30%,#D4AF37 48%,#B8860B 62%,#D4AF37 75%,#FFD700 85%,#FFE55C 92%,#FFFACD 100%)!important;-webkit-background-clip:text!important;-webkit-text-fill-color:transparent!important;background-clip:text!important;font-weight:800!important;line-height:1.15!important;filter:drop-shadow(0 3px 10px rgba(212,175,55,0.5))!important;background-color:transparent!important;letter-spacing:-0.02em!important}
h1{font-size:2.4em!important;margin-bottom:0.3em!important}h2{font-size:1.8em!important;border-bottom:3px solid rgba(212,175,55,0.4)!important;padding-bottom:0.3em!important;margin:1.2em 0 0.4em!important;box-shadow:0 3px 0 rgba(184,134,11,0.2)!important}h3{font-size:1.4em!important;margin:1em 0 0.3em!important}h4{font-size:1.15em!important}h5{font-size:1.05em!important}h6{font-size:0.95em!important}
p,li,td,th,div,label,figcaption,cite,small,time,dd,dt,dl,span{color:#D5C9B5!important;background-color:transparent!important}
strong,b,em{background:linear-gradient(135deg,#FFFFFF,#FFFACD,#FFE55C,#FFD700,#D4AF37)!important;-webkit-background-clip:text!important;-webkit-text-fill-color:transparent!important;background-clip:text!important;font-weight:800!important;background-color:transparent!important;filter:drop-shadow(0 2px 4px rgba(212,175,55,0.4))!important;letter-spacing:0.02em!important}
a,a:link,a:visited,a:active{color:#D4AF37!important;text-decoration:none!important;border-bottom:2px solid rgba(212,175,55,0.3)!important;transition:all 0.4s ease!important;padding-bottom:2px!important}a:hover{color:#FFD700!important;border-bottom-color:rgba(255,215,0,0.8)!important;text-shadow:0 0 20px rgba(255,215,0,0.5),0 0 40px rgba(255,215,0,0.2)!important}
.entry-content{max-width:700px!important;margin:0 auto!important;padding:2.5em 1.5em!important}
.entry-content p{margin-bottom:1.3em!important;font-size:1em!important;line-height:1.8!important}
.entry-content blockquote{border-left:5px solid #FFD700!important;padding:1.2em 1.5em!important;margin:2em 0!important;background:linear-gradient(135deg,rgba(212,175,55,0.12),rgba(212,175,55,0.03))!important;background-color:rgba(212,175,55,0.07)!important;border-radius:0 16px 16px 0!important;color:#C5B89E!important;font-style:italic!important;font-size:1em!important;line-height:1.7!important;box-shadow:inset 0 0 30px rgba(212,175,55,0.05),8px 0 25px rgba(0,0,0,0.3)!important}
.entry-content blockquote *{color:#C5B89E!important}
.entry-content pre{background:#020408!important;background-color:#020408!important;color:#D8CDB8!important;padding:1.5em!important;border-radius:16px!important;border:2px solid rgba(212,175,55,0.25)!important;box-shadow:inset 0 3px 12px rgba(0,0,0,0.5),0 8px 25px rgba(0,0,0,0.4)!important;font-size:0.85em!important;line-height:1.6!important}
.entry-content code{background:rgba(212,175,55,0.15)!important;background-color:rgba(212,175,55,0.15)!important;color:#FFD700!important;padding:0.2em 0.6em!important;border-radius:10px!important;border:1px solid rgba(212,175,55,0.25)!important;font-weight:800!important;font-size:0.82em!important}
.entry-content pre code{background:transparent!important;background-color:transparent!important;color:#D8CDB8!important;border:none!important;font-weight:400!important}
img{max-width:100%!important;height:auto!important;border-radius:16px!important;box-shadow:0 16px 48px rgba(0,0,0,0.8),0 0 0 2px rgba(212,175,55,0.2)!important}
table{width:100%!important;border-collapse:collapse!important;border-radius:16px!important;overflow:hidden!important;box-shadow:0 12px 36px rgba(0,0,0,0.7),0 0 0 2px rgba(212,175,55,0.15)!important}
th{background:linear-gradient(180deg,#1E2838,#141C2C)!important;background-color:#1E2838!important;color:#FFD700!important;padding:1em 1.2em!important;font-weight:800!important;border-bottom:3px solid rgba(212,175,55,0.6)!important;text-transform:uppercase!important;font-size:0.7em!important;letter-spacing:0.2em!important}
td{padding:1em 1.2em!important;border-bottom:1px solid rgba(212,175,55,0.12)!important;color:#CDC0AA!important;font-size:0.95em!important;line-height:1.6!important}
tr:nth-child(even){background:rgba(212,175,55,0.05)!important;background-color:rgba(212,175,55,0.05)!important}
tr:hover{background:rgba(212,175,55,0.1)!important;background-color:rgba(212,175,55,0.1)!important}
.wp-block-post{background:linear-gradient(160deg,#1A2234 0%,#121828 40%,#0E1420 70%,#0A0E18 100%)!important;background-color:#1A2234!important;border:2px solid rgba(212,175,55,0.18)!important;border-radius:20px!important;padding:28px!important;margin-bottom:2em!important;box-shadow:0 14px 40px rgba(0,0,0,0.7),0 0 0 2px rgba(212,175,55,0.1),inset 0 1px 0 rgba(255,255,255,0.05)!important;transition:border-color 0.5s ease,box-shadow 0.5s ease!important;position:relative!important;overflow:hidden!important}
.wp-block-post::before{content:""!important;position:absolute!important;top:0!important;left:0!important;right:0!important;height:3px!important;background:linear-gradient(90deg,transparent,rgba(212,175,55,0.4),rgba(255,215,0,0.6),rgba(212,175,55,0.4),transparent)!important}
.wp-block-post:hover{border-color:rgba(212,175,55,0.55)!important;box-shadow:0 20px 60px rgba(0,0,0,0.8),0 0 80px rgba(212,175,55,0.15),0 0 0 3px rgba(212,175,55,0.25)!important}
.wp-block-post-title a{color:#FFFFFF!important;font-weight:800!important;font-size:1.1em!important;letter-spacing:-0.01em!important;line-height:1.3!important}.wp-block-post-title a:hover{color:#FFD700!important;text-shadow:0 0 20px rgba(255,215,0,0.4)!important}
.wp-block-button__link{background:linear-gradient(135deg,#FFFFFF 0%,#FFFACD 10%,#FFE55C 22%,#FFD700 38%,#D4AF37 55%,#B8860B 70%,#D4AF37 82%,#FFD700 92%,#FFE55C 100%)!important;background-color:#D4AF37!important;color:#0A0A0A!important;border-radius:10px!important;font-weight:900!important;text-transform:uppercase!important;letter-spacing:0.12em!important;font-size:0.72em!important;box-shadow:0 8px 24px rgba(212,175,55,0.6),0 0 0 2px rgba(255,215,0,0.5)!important;padding:12px 28px!important;border:2px solid rgba(255,215,0,0.6)!important}
.wp-block-button__link:hover{box-shadow:0 14px 40px rgba(212,175,55,0.8),0 0 60px rgba(212,175,55,0.3),0 0 0 3px rgba(255,215,0,0.7)!important}
.wp-block-separator{border-color:rgba(212,175,55,0.4)!important;margin:3em 0!important;box-shadow:0 0 30px rgba(212,175,55,0.2)!important;height:3px!important}
header.wp-block-template-part{background:linear-gradient(180deg,#121828,#0A0E18)!important;background-color:#121828!important;border-bottom:3px solid rgba(212,175,55,0.3)!important;box-shadow:0 12px 36px rgba(0,0,0,0.8)!important;padding:0.8em 0!important}
.wp-block-site-title a{background:linear-gradient(135deg,#FFFFFF,#FFFACD,#FFE55C,#FFD700,#D4AF37)!important;-webkit-background-clip:text!important;-webkit-text-fill-color:transparent!important;background-clip:text!important;font-weight:900!important;font-size:1.3em!important;letter-spacing:0.12em!important;text-transform:uppercase!important;background-color:transparent!important;filter:drop-shadow(0 3px 6px rgba(212,175,55,0.5))!important}
.wp-block-navigation a{color:#B0A898!important;font-weight:700!important;text-transform:uppercase!important;font-size:0.68em!important;letter-spacing:0.18em!important;border-bottom:none!important;padding:0.5em 1em!important;border-radius:8px!important;transition:all 0.4s ease!important}.wp-block-navigation a:hover{color:#FFD700!important;background:rgba(212,175,55,0.12)!important;text-shadow:0 0 20px rgba(255,215,0,0.5)!important;box-shadow:0 0 20px rgba(212,175,55,0.15)!important}
footer.wp-block-template-part{background:linear-gradient(180deg,#0A0E18,#040608,#010102)!important;background-color:#0A0E18!important;border-top:3px solid rgba(212,175,55,0.3)!important;padding:2.5em 0 2em!important;box-shadow:0 -12px 36px rgba(0,0,0,0.7)!important}
footer.wp-block-template-part *{color:#A89E90!important;font-weight:500!important}footer.wp-block-template-part a{color:#D4AF37!important;border-bottom:none!important;font-weight:700!important}footer.wp-block-template-part a:hover{color:#FFD700!important;text-shadow:0 0 20px rgba(255,215,0,0.6)!important}
footer div[style*=fefcf5]{background:#0A0E18!important;background-color:#0A0E18!important;border-top:1px solid rgba(212,175,55,0.15)!important}
footer div[style*=fefcf5] *{color:#4A453D!important}footer div[style*=fefcf5] a{color:#6B6358!important;border-bottom:none!important}footer div[style*=fefcf5] a:hover{color:#8B7355!important}
.widget-area,.sidebar{background:linear-gradient(160deg,#1A2234,#121828)!important;background-color:#1A2234!important;border:2px solid rgba(212,175,55,0.18)!important;border-radius:20px!important;padding:1.5em!important;box-shadow:0 12px 36px rgba(0,0,0,0.6)!important}
.widget-title{color:#FFD700!important;font-weight:800!important;text-transform:uppercase!important;letter-spacing:0.15em!important;font-size:0.72em!important;border-bottom:3px solid rgba(212,175,55,0.4)!important;padding-bottom:0.8em!important;margin-bottom:1em!important}
.widget-area a{color:#D4AF37!important;border-bottom:none!important;font-weight:700!important}.widget-area a:hover{color:#FFD700!important;text-shadow:0 0 15px rgba(255,215,0,0.5)!important}
.wp-block-latest-posts__post-title a{color:#FFF!important;border-bottom:none!important;font-weight:800!important;font-size:1em!important}.wp-block-latest-posts__post-title a:hover{color:#FFD700!important}
.wp-block-cover{border-radius:20px!important;overflow:hidden!important;box-shadow:0 24px 72px rgba(0,0,0,0.9),0 0 0 3px rgba(212,175,55,0.2)!important}
.wp-block-categories a,.wp-block-archives a,.wp-block-latest-posts a,.wp-block-tag-cloud a{color:#D4AF37!important;border-bottom:none!important;font-weight:700!important}
.wp-block-categories a:hover,.wp-block-archives a:hover,.wp-block-latest-posts a:hover,.wp-block-tag-cloud a:hover{color:#FFD700!important}
.wp-block-categories__count{background:linear-gradient(135deg,rgba(212,175,55,0.2),rgba(212,175,55,0.06))!important;background-color:rgba(212,175,55,0.12)!important;color:#FFD700!important;border-radius:20px!important;padding:0.25em 0.9em!important;font-size:0.72em!important;font-weight:800!important;border:1px solid rgba(212,175,55,0.25)!important}
blockquote,.wp-block-quote,.wp-block-pullquote{border-left:5px solid #FFD700!important;background:linear-gradient(135deg,rgba(212,175,55,0.12),rgba(212,175,55,0.03))!important;background-color:rgba(212,175,55,0.07)!important;border-radius:0 16px 16px 0!important}
blockquote *,.wp-block-quote *,.wp-block-pullquote *{color:#C5B89E!important}
.has-white-background-color,.has-base-background-color{background-color:#1A2234!important}
.has-white-color,.has-base-color{color:#D5C9B5!important}
a:where(:not(.wp-element-button)){color:#D4AF37!important;border-bottom:2px solid rgba(212,175,55,0.25)!important}a:where(:not(.wp-element-button)):hover{color:#FFD700!important}
.wp-block-tag-cloud a{background:linear-gradient(135deg,rgba(212,175,55,0.15),rgba(212,175,55,0.05))!important;background-color:rgba(212,175,55,0.1)!important;color:#D4AF37!important;border-radius:20px!important;padding:0.4em 0.9em!important;border:1px solid rgba(212,175,55,0.2)!important;font-weight:700!important;font-size:0.82em!important}
.wp-block-tag-cloud a:hover{background:rgba(212,175,55,0.25)!important;background-color:rgba(212,175,55,0.2)!important;color:#FFD700!important;border-color:rgba(212,175,55,0.4)!important}
.wp-block-media-text{border-radius:20px!important;overflow:hidden!important;box-shadow:0 16px 48px rgba(0,0,0,0.8)!important}
.wp-block-gallery .wp-block-image{border-radius:16px!important;overflow:hidden!important}
figcaption{color:#A89E90!important;font-style:italic!important;font-size:0.85em!important}
@media(max-width:768px){.entry-content{padding:1.5em 1em!important}h1{font-size:2em!important}h2{font-size:1.5em!important}h3{font-size:1.2em!important}.wp-block-post{padding:20px!important;margin-bottom:1.5em!important;border-radius:16px!important}.wp-block-button__link{padding:10px 24px!important;font-size:0.75em!important}}
');
}, 99);

// Keyframes
add_action('wp_head', function() {
    echo '<style>
    @keyframes scrollBanner{0%{transform:translateX(0)}100%{transform:translateX(-50%)}}
    @keyframes shimmerSlide{0%{transform:translateX(-100%)}100%{transform:translateX(200%)}}
    @keyframes pulseWord{0%{opacity:0.65;text-shadow:0 0 8px rgba(255,215,0,0.4)}100%{opacity:1;text-shadow:0 0 30px rgba(255,215,0,1),0 0 60px rgba(255,215,0,0.5),0 0 90px rgba(255,215,0,0.2)}}
    @keyframes pulseDot{0%,100%{opacity:0.1;transform:scale(0.4)}50%{opacity:1;transform:scale(2)}}
    </style>';
}, 1);

// Banner function
function ai_money_machine_banner() {
    $emojis = array('💰','💎','👑','🏆','✨','💫','🌟','⭐','🪙','💍','🔱','⚜️','🦅','🦁','🏰','🗡️','🫅','👸','🤴','💃');
    $words = array('EXCLUSIVE','LUXURY','PREMIUM','ELITE','FORTUNE','EXCELLENCE','DIAMOND','SOVEREIGN','WEALTH','OPULENCE','MAJESTY','GRANDEUR','PRESTIGE','IMPERIAL','SUPREME','REGAL','MAGNIFICENT','SPLENDID','GLORIOUS','TRIUMPHANT','TRANSCENDENT','EPIC','LEGENDARY','MYTHICAL');
    echo '<div style="background:linear-gradient(90deg,#020408,#081018,#0E1828,#141E30,#0E1828,#081018,#020408);border-top:3px solid #FFD700;border-bottom:3px solid #FFD700;padding:0.8em 0;overflow:hidden;position:relative;width:100%;margin:0;box-shadow:0 0 50px rgba(212,175,55,0.3),inset 0 0 30px rgba(212,175,55,0.05)">';
    echo '<div style="position:absolute;top:0;left:0;right:0;bottom:0;background:linear-gradient(90deg,transparent,rgba(255,215,0,0.15),transparent);animation:shimmerSlide 5s ease-in-out infinite;pointer-events:none;z-index:0"></div>';
    echo '<div style="display:flex;animation:scrollBanner 60s linear infinite;white-space:nowrap;position:relative;z-index:1;width:max-content">';
    for ($i = 0; $i < 5; $i++) {
        $idx = 0;
        foreach ($words as $w) {
            $emoji = $emojis[$idx % count($emojis)];
            $delay = $idx * 0.14;
            $dotDelay = $idx * 0.08;
            echo '<span style="color:#FFD700;font-size:0.9em;font-weight:900;letter-spacing:0.4em;text-transform:uppercase;padding:0 0.25em;text-shadow:0 0 12px rgba(255,215,0,0.7),0 0 25px rgba(255,215,0,0.3);display:inline-block;animation:pulseWord 4.5s ease-in-out infinite alternate;animation-delay:'.$delay.'s">'.$emoji.' '.$w.'</span>';
            echo '<span style="color:#B8860B;padding:0 0.06em;font-size:0.25em;display:inline-block;animation:pulseDot 2.5s ease-in-out infinite;animation-delay:'.$dotDelay.'s;text-shadow:0 0 10px rgba(184,134,11,1)">&#9670;</span>';
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
print("Done - COMPACT WEALTH")
