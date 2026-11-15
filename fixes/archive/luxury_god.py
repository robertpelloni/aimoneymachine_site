import subprocess

with open(
    "/var/www/aimoneymachine/wp-content/themes/twentytwentyfive/functions.php"
) as f:
    php = f.read()

marker = "// AI Money Machine"
if marker in php:
    php = php[: php.index(marker)]

php += r"""
// AI Money Machine - GOD MODE WEALTH
add_action('wp_enqueue_scripts', function() {
    wp_add_inline_style('twentytwentyfive-style', '
*,*::before,*::after{background-color:transparent!important;background-image:none!important}
body{font-family:Palatino Linotype,Book Antiqua,Palatino,Georgia,serif!important;font-size:19px!important;line-height:2!important;color:#EEE4D4!important;background:linear-gradient(180deg,#000000 0%,#020304 15%,#050810 35%,#080C18 55%,#0B1020 75%,#0E1428 100%)!important;background-color:#000000!important}
h1,h2,h3,h4,h5,h6{background:linear-gradient(135deg,#FFFFFF 0%,#FFFFF0 5%,#FFFACD 12%,#FFE88C 20%,#FFD700 30%,#DAA520 42%,#D4AF37 52%,#B8860B 62%,#D4AF37 72%,#DAA520 80%,#FFD700 88%,#FFE88C 94%,#FFFACD 100%)!important;-webkit-background-clip:text!important;-webkit-text-fill-color:transparent!important;background-clip:text!important;font-weight:900!important;line-height:1.05!important;filter:drop-shadow(0 6px 25px rgba(212,175,55,0.8))!important;background-color:transparent!important;letter-spacing:-0.05em!important}
h1{font-size:5em!important;margin-bottom:0.15em!important}h2{font-size:3.2em!important;border-bottom:5px solid rgba(212,175,55,0.5)!important;padding-bottom:0.4em!important;margin:1.5em 0 0.5em!important;box-shadow:0 5px 0 rgba(184,134,11,0.3)!important}h3{font-size:2.4em!important;margin:1.2em 0 0.4em!important}h4{font-size:1.6em!important}h5{font-size:1.3em!important}h6{font-size:1.1em!important}
p,li,td,th,div,label,figcaption,cite,small,time,dd,dt,dl,span{color:#E0D5C0!important;background-color:transparent!important}
strong,b,em{background:linear-gradient(135deg,#FFFFFF,#FFFFF0,#FFFACD,#FFE88C,#FFD700,#DAA520,#D4AF37)!important;-webkit-background-clip:text!important;-webkit-text-fill-color:transparent!important;background-clip:text!important;font-weight:900!important;background-color:transparent!important;filter:drop-shadow(0 4px 8px rgba(212,175,55,0.6))!important;letter-spacing:0.04em!important}
a,a:link,a:visited,a:active{color:#DAA520!important;text-decoration:none!important;border-bottom:3px solid rgba(212,175,55,0.35)!important;transition:all 0.6s ease!important;padding-bottom:4px!important}a:hover{color:#FFD700!important;border-bottom-color:rgba(255,215,0,0.9)!important;text-shadow:0 0 40px rgba(255,215,0,0.7),0 0 80px rgba(255,215,0,0.3),0 0 120px rgba(255,215,0,0.1)!important}
.entry-content{max-width:680px!important;margin:0 auto!important;padding:5em 3em!important}
.entry-content p{margin-bottom:2em!important;font-size:1.1em!important;line-height:2.1!important}
.entry-content blockquote{border-left:8px solid #FFD700!important;padding:2.5em 3em!important;margin:3.5em 0!important;background:linear-gradient(135deg,rgba(212,175,55,0.15),rgba(212,175,55,0.04))!important;background-color:rgba(212,175,55,0.08)!important;border-radius:0 28px 28px 0!important;color:#D0C4AA!important;font-style:italic!important;font-size:1.2em!important;line-height:1.9!important;box-shadow:inset 0 0 60px rgba(212,175,55,0.08),16px 0 50px rgba(0,0,0,0.5)!important}
.entry-content blockquote *{color:#D0C4AA!important}
.entry-content pre{background:#020408!important;background-color:#020408!important;color:#E0D5C0!important;padding:3em!important;border-radius:28px!important;border:3px solid rgba(212,175,55,0.3)!important;box-shadow:inset 0 5px 25px rgba(0,0,0,0.7),0 16px 50px rgba(0,0,0,0.6)!important;font-size:0.9em!important;line-height:1.8!important}
.entry-content code{background:rgba(212,175,55,0.18)!important;background-color:rgba(212,175,55,0.18)!important;color:#FFD700!important;padding:0.35em 1em!important;border-radius:14px!important;border:2px solid rgba(212,175,55,0.3)!important;font-weight:900!important;font-size:0.85em!important}
.entry-content pre code{background:transparent!important;background-color:transparent!important;color:#E0D5C0!important;border:none!important;font-weight:400!important}
img{max-width:100%!important;height:auto!important;border-radius:28px!important;box-shadow:0 32px 96px rgba(0,0,0,1),0 0 0 4px rgba(212,175,55,0.25)!important}
table{width:100%!important;border-collapse:collapse!important;border-radius:28px!important;overflow:hidden!important;box-shadow:0 24px 72px rgba(0,0,0,0.8),0 0 0 4px rgba(212,175,55,0.2)!important}
th{background:linear-gradient(180deg,#222E42,#182030)!important;background-color:#222E42!important;color:#FFD700!important;padding:1.8em 2em!important;font-weight:900!important;border-bottom:5px solid rgba(212,175,55,0.7)!important;text-transform:uppercase!important;font-size:0.72em!important;letter-spacing:0.3em!important}
td{padding:1.8em 2em!important;border-bottom:2px solid rgba(212,175,55,0.15)!important;color:#D5C8B0!important;font-size:1.08em!important;line-height:1.8!important}
tr:nth-child(even){background:rgba(212,175,55,0.06)!important;background-color:rgba(212,175,55,0.06)!important}
tr:hover{background:rgba(212,175,55,0.12)!important;background-color:rgba(212,175,55,0.12)!important}
.wp-block-post{background:linear-gradient(160deg,#1E2A40 0%,#162030 30%,#101828 60%,#0C1020 100%)!important;background-color:#1E2A40!important;border:3px solid rgba(212,175,55,0.2)!important;border-radius:36px!important;padding:56px!important;margin-bottom:3.5em!important;box-shadow:0 28px 84px rgba(0,0,0,0.8),0 0 0 3px rgba(212,175,55,0.12),inset 0 1px 0 rgba(255,255,255,0.06)!important;transition:border-color 0.7s ease,box-shadow 0.7s ease!important;position:relative!important;overflow:hidden!important}
.wp-block-post::before{content:""!important;position:absolute!important;top:0!important;left:0!important;right:0!important;height:4px!important;background:linear-gradient(90deg,transparent,rgba(212,175,55,0.5),rgba(255,215,0,0.8),rgba(212,175,55,0.5),transparent)!important}
.wp-block-post:hover{border-color:rgba(212,175,55,0.6)!important;box-shadow:0 40px 120px rgba(0,0,0,0.9),0 0 150px rgba(212,175,55,0.2),0 0 0 4px rgba(212,175,55,0.3)!important}
.wp-block-post-title a{color:#FFFFFF!important;font-weight:900!important;font-size:1.3em!important;letter-spacing:-0.03em!important;line-height:1.2!important}.wp-block-post-title a:hover{color:#FFD700!important;text-shadow:0 0 40px rgba(255,215,0,0.6)!important}
.wp-block-button__link{background:linear-gradient(135deg,#FFFFFF 0%,#FFFFF0 8%,#FFFACD 16%,#FFE88C 26%,#FFD700 38%,#DAA520 50%,#D4AF37 60%,#B8860B 72%,#D4AF37 82%,#DAA520 90%,#FFD700 100%)!important;background-color:#D4AF37!important;color:#0A0A0A!important;border-radius:18px!important;font-weight:900!important;text-transform:uppercase!important;letter-spacing:0.18em!important;font-size:0.75em!important;box-shadow:0 18px 50px rgba(212,175,55,0.7),0 0 0 4px rgba(255,215,0,0.6)!important;padding:24px 56px!important;border:4px solid rgba(255,215,0,0.7)!important}
.wp-block-button__link:hover{box-shadow:0 28px 84px rgba(212,175,55,0.9),0 0 120px rgba(212,175,55,0.4),0 0 0 5px rgba(255,215,0,0.8)!important}
.wp-block-separator{border-color:rgba(212,175,55,0.5)!important;margin:5em 0!important;box-shadow:0 0 50px rgba(212,175,55,0.25)!important;height:4px!important}
header.wp-block-template-part{background:linear-gradient(180deg,#162030,#0C1020)!important;background-color:#162030!important;border-bottom:4px solid rgba(212,175,55,0.35)!important;box-shadow:0 20px 60px rgba(0,0,0,0.9)!important;padding:1.6em 0!important}
.wp-block-site-title a{background:linear-gradient(135deg,#FFFFFF,#FFFFF0,#FFFACD,#FFE88C,#FFD700,#DAA520,#D4AF37)!important;-webkit-background-clip:text!important;-webkit-text-fill-color:transparent!important;background-clip:text!important;font-weight:900!important;font-size:2em!important;letter-spacing:0.18em!important;text-transform:uppercase!important;background-color:transparent!important;filter:drop-shadow(0 5px 10px rgba(212,175,55,0.6))!important}
.wp-block-navigation a{color:#C0B8A8!important;font-weight:900!important;text-transform:uppercase!important;font-size:0.7em!important;letter-spacing:0.25em!important;border-bottom:none!important;padding:0.8em 1.6em!important;border-radius:12px!important;transition:all 0.5s ease!important}.wp-block-navigation a:hover{color:#FFD700!important;background:rgba(212,175,55,0.15)!important;text-shadow:0 0 30px rgba(255,215,0,0.6)!important;box-shadow:0 0 30px rgba(212,175,55,0.2)!important}
footer.wp-block-template-part{background:linear-gradient(180deg,#0C1020,#060810,#020304,#000000)!important;background-color:#0C1020!important;border-top:4px solid rgba(212,175,55,0.35)!important;padding:5em 0 3.5em!important;box-shadow:0 -20px 60px rgba(0,0,0,0.8)!important}
footer.wp-block-template-part *{color:#B8AEA0!important;font-weight:600!important}footer.wp-block-template-part a{color:#DAA520!important;border-bottom:none!important;font-weight:800!important}footer.wp-block-template-part a:hover{color:#FFD700!important;text-shadow:0 0 30px rgba(255,215,0,0.7)!important}
footer div[style*=fefcf5]{background:#0C1020!important;background-color:#0C1020!important;border-top:2px solid rgba(212,175,55,0.2)!important}
footer div[style*=fefcf5] *{color:#4A453D!important}footer div[style*=fefcf5] a{color:#6B6358!important;border-bottom:none!important}footer div[style*=fefcf5] a:hover{color:#8B7355!important}
.widget-area,.sidebar{background:linear-gradient(160deg,#1E2A40,#162030)!important;background-color:#1E2A40!important;border:3px solid rgba(212,175,55,0.2)!important;border-radius:32px!important;padding:3em!important;box-shadow:0 20px 60px rgba(0,0,0,0.7)!important}
.widget-title{color:#FFD700!important;font-weight:900!important;text-transform:uppercase!important;letter-spacing:0.2em!important;font-size:0.75em!important;border-bottom:4px solid rgba(212,175,55,0.5)!important;padding-bottom:1.5em!important;margin-bottom:2em!important}
.widget-area a{color:#DAA520!important;border-bottom:none!important;font-weight:800!important}.widget-area a:hover{color:#FFD700!important;text-shadow:0 0 25px rgba(255,215,0,0.6)!important}
.wp-block-latest-posts__post-title a{color:#FFF!important;border-bottom:none!important;font-weight:900!important;font-size:1.15em!important}.wp-block-latest-posts__post-title a:hover{color:#FFD700!important}
.wp-block-cover{border-radius:32px!important;overflow:hidden!important;box-shadow:0 40px 120px rgba(0,0,0,1),0 0 0 4px rgba(212,175,55,0.25)!important}
.wp-block-categories a,.wp-block-archives a,.wp-block-latest-posts a,.wp-block-tag-cloud a{color:#DAA520!important;border-bottom:none!important;font-weight:800!important}
.wp-block-categories a:hover,.wp-block-archives a:hover,.wp-block-latest-posts a:hover,.wp-block-tag-cloud a:hover{color:#FFD700!important}
.wp-block-categories__count{background:linear-gradient(135deg,rgba(212,175,55,0.25),rgba(212,175,55,0.08))!important;background-color:rgba(212,175,55,0.15)!important;color:#FFD700!important;border-radius:32px!important;padding:0.4em 1.4em!important;font-size:0.75em!important;font-weight:900!important;border:2px solid rgba(212,175,55,0.3)!important}
blockquote,.wp-block-quote,.wp-block-pullquote{border-left:8px solid #FFD700!important;background:linear-gradient(135deg,rgba(212,175,55,0.15),rgba(212,175,55,0.04))!important;background-color:rgba(212,175,55,0.08)!important;border-radius:0 28px 28px 0!important}
blockquote *,.wp-block-quote *,.wp-block-pullquote *{color:#D0C4AA!important}
.has-white-background-color,.has-base-background-color{background-color:#1E2A40!important}
.has-white-color,.has-base-color{color:#E0D5C0!important}
a:where(:not(.wp-element-button)){color:#DAA520!important;border-bottom:3px solid rgba(212,175,55,0.3)!important}a:where(:not(.wp-element-button)):hover{color:#FFD700!important}
.wp-block-tag-cloud a{background:linear-gradient(135deg,rgba(212,175,55,0.2),rgba(212,175,55,0.06))!important;background-color:rgba(212,175,55,0.12)!important;color:#DAA520!important;border-radius:32px!important;padding:0.6em 1.4em!important;border:2px solid rgba(212,175,55,0.25)!important;font-weight:800!important;font-size:0.88em!important}
.wp-block-tag-cloud a:hover{background:rgba(212,175,55,0.3)!important;background-color:rgba(212,175,55,0.25)!important;color:#FFD700!important;border-color:rgba(212,175,55,0.5)!important}
.wp-block-media-text{border-radius:28px!important;overflow:hidden!important;box-shadow:0 28px 84px rgba(0,0,0,0.9)!important}
.wp-block-gallery .wp-block-image{border-radius:24px!important;overflow:hidden!important}
figcaption{color:#B8AEA0!important;font-style:italic!important;font-size:0.88em!important}
.wp-block-comments-query-loop{border-radius:24px!important;border:2px solid rgba(212,175,55,0.15)!important;padding:2em!important}
.comment-body{border-bottom:1px solid rgba(212,175,55,0.1)!important;padding:1.5em 0!important}
@media(max-width:768px){.entry-content{padding:2.5em 1.8em!important}h1{font-size:3em!important}h2{font-size:2.4em!important}h3{font-size:1.8em!important}.wp-block-post{padding:36px!important;margin-bottom:2.5em!important;border-radius:28px!important}.wp-block-button__link{padding:20px 40px!important;font-size:0.8em!important}}
');
}, 99);

// Keyframes
add_action('wp_head', function() {
    echo '<style>
    @keyframes scrollBanner{0%{transform:translateX(0)}100%{transform:translateX(-50%)}}
    @keyframes shimmerSlide{0%{transform:translateX(-100%)}100%{transform:translateX(200%)}}
    @keyframes pulseWord{0%{opacity:0.65;text-shadow:0 0 10px rgba(255,215,0,0.4)}100%{opacity:1;text-shadow:0 0 40px rgba(255,215,0,1),0 0 80px rgba(255,215,0,0.6),0 0 120px rgba(255,215,0,0.25)}}
    @keyframes pulseDot{0%,100%{opacity:0.08;transform:scale(0.3)}50%{opacity:1;transform:scale(2.5)}}
    </style>';
}, 1);

// Banner function
function ai_money_machine_banner() {
    $emojis = array('💰','💎','👑','🏆','✨','💫','🌟','⭐','🪙','💍','🔱','⚜️','🦅','🦁','🏰','🗡️','🫅','👸','🤴','💃');
    $words = array('EXCLUSIVE','LUXURY','PREMIUM','ELITE','FORTUNE','EXCELLENCE','DIAMOND','SOVEREIGN','WEALTH','OPULENCE','MAJESTY','GRANDEUR','PRESTIGE','IMPERIAL','SUPREME','REGAL','MAGNIFICENT','SPLENDID','GLORIOUS','TRIUMPHANT','TRANSCENDENT','EPIC','LEGENDARY','MYTHICAL');
    echo '<div style="background:linear-gradient(90deg,#020408,#081018,#0E1828,#141E30,#0E1828,#081018,#020408);border-top:4px solid #FFD700;border-bottom:4px solid #FFD700;padding:1.5em 0;overflow:hidden;position:relative;width:100%;margin:0;box-shadow:0 0 100px rgba(212,175,55,0.35),inset 0 0 60px rgba(212,175,55,0.08)">';
    echo '<div style="position:absolute;top:0;left:0;right:0;bottom:0;background:linear-gradient(90deg,transparent,rgba(255,215,0,0.2),transparent);animation:shimmerSlide 5s ease-in-out infinite;pointer-events:none;z-index:0"></div>';
    echo '<div style="display:flex;animation:scrollBanner 60s linear infinite;white-space:nowrap;position:relative;z-index:1;width:max-content">';
    for ($i = 0; $i < 5; $i++) {
        $idx = 0;
        foreach ($words as $w) {
            $emoji = $emojis[$idx % count($emojis)];
            $delay = $idx * 0.14;
            $dotDelay = $idx * 0.08;
            echo '<span style="color:#FFD700;font-size:1.35em;font-weight:900;letter-spacing:0.5em;text-transform:uppercase;padding:0 0.3em;text-shadow:0 0 20px rgba(255,215,0,0.9),0 0 40px rgba(255,215,0,0.4);display:inline-block;animation:pulseWord 4.5s ease-in-out infinite alternate;animation-delay:'.$delay.'s">'.$emoji.' '.$w.'</span>';
            echo '<span style="color:#B8860B;padding:0 0.08em;font-size:0.3em;display:inline-block;animation:pulseDot 2.5s ease-in-out infinite;animation-delay:'.$dotDelay.'s;text-shadow:0 0 15px rgba(184,134,11,1)">&#9670;</span>';
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
print("Done - GOD MODE WEALTH")
