<?php
/**
 * Plugin Name: Affiliate Links Manager
 * Description: Automatically adds affiliate links to content
 */

// Affiliate link database
function get_affiliate_links() {
    return array(
        // AI Tools
        'chatgpt' => array('url' => 'https://openai.com/chatgpt/pricing/', 'text' => 'ChatGPT Plus', 'commission' => '$20/mo'),
        'claude' => array('url' => 'https://claude.ai/pricing', 'text' => 'Claude Pro', 'commission' => '$20/mo'),
        'midjourney' => array('url' => 'https://www.midjourney.com/', 'text' => 'Midjourney', 'commission' => '20%'),
        'jasper' => array('url' => 'https://www.jasper.ai/', 'text' => 'Jasper AI', 'commission' => '30%'),
        'copy.ai' => array('url' => 'https://www.copy.ai/', 'text' => 'Copy.ai', 'commission' => '45%'),
        'writesonic' => array('url' => 'https://writesonic.com/', 'text' => 'Writesonic', 'commission' => '30%'),
        'grammarly' => array('url' => 'https://www.grammarly.com/', 'text' => 'Grammarly', 'commission' => '$20/sale'),
        'surfer' => array('url' => 'https://surferseo.com/', 'text' => 'Surfer SEO', 'commission' => '25%'),
        
        // Crypto Exchanges
        'binance' => array('url' => 'https://www.binance.com/en/register', 'text' => 'Binance', 'commission' => '50% trading fees'),
        'coinbase' => array('url' => 'https://www.coinbase.com/', 'text' => 'Coinbase', 'commission' => '$10/signup'),
        'kraken' => array('url' => 'https://www.kraken.com/', 'text' => 'Kraken', 'commission' => '20% trading fees'),
        'crypto.com' => array('url' => 'https://crypto.com/', 'text' => 'Crypto.com', 'commission' => '$25/signup'),
        
        // Hosting
        'hetzner' => array('url' => 'https://www.hetzner.com/cloud', 'text' => 'Hetzner Cloud', 'commission' => '€20/signup'),
        'digitalocean' => array('url' => 'https://www.digitalocean.com/', 'text' => 'DigitalOcean', 'commission' => '$200 credit'),
        'aws' => array('url' => 'https://aws.amazon.com/free/', 'text' => 'AWS Free Tier', 'commission' => 'varies'),
        'cloudflare' => array('url' => 'https://www.cloudflare.com/', 'text' => 'Cloudflare', 'commission' => 'varies'),
        
        // Courses
        'udemy' => array('url' => 'https://www.udemy.com/', 'text' => 'Udemy', 'commission' => '15%'),
        'coursera' => array('url' => 'https://www.coursera.org/', 'text' => 'Coursera', 'commission' => '10-45%'),
        'skillshare' => array('url' => 'https://www.skillshare.com/', 'text' => 'Skillshare', 'commission' => '$7/trial'),
        
        // Productivity
        'notion' => array('url' => 'https://www.notion.so/', 'text' => 'Notion', 'commission' => '$10/signup'),
        'zapier' => array('url' => 'https://zapier.com/', 'text' => 'Zapier', 'commission' => 'varies'),
        'make' => array('url' => 'https://www.make.com/', 'text' => 'Make (Integromat)', 'commission' => 'varies'),
    );
}

// Add affiliate links to post content
function add_affiliate_links_to_content($content) {
    if (!is_single()) return $content;
    
    $links = get_affiliate_links();
    $added = array();
    
    // Find relevant keywords in content and add links
    $keywords = array(
        'ChatGPT' => 'chatgpt',
        'Claude' => 'claude',
        'Midjourney' => 'midjourney',
        'Jasper' => 'jasper',
        'Copy.ai' => 'copy.ai',
        'Writesonic' => 'writesonic',
        'Grammarly' => 'grammarly',
        'Surfer SEO' => 'surfer',
        'Binance' => 'binance',
        'Coinbase' => 'coinbase',
        'Kraken' => 'kraken',
        'Crypto.com' => 'crypto.com',
        'Hetzner' => 'hetzner',
        'DigitalOcean' => 'digitalocean',
        'AWS' => 'aws',
        'Cloudflare' => 'cloudflare',
        'Udemy' => 'udemy',
        'Coursera' => 'coursera',
        'Skillshare' => 'skillshare',
        'Notion' => 'notion',
        'Zapier' => 'zapier',
        'Make' => 'make',
    );
    
    foreach ($keywords as $keyword => $link_key) {
        if (in_array($link_key, $added)) continue;
        if (stripos($content, $keyword) !== false) {
            $link = $links[$link_key];
            // Add link to first occurrence only
            $linked = '<a href="' . esc_url($link['url']) . '?ref=aimoneymachine" target="_blank" rel="nofollow noopener sponsored">' . $keyword . '</a>';
            $content = preg_replace('/\b' . preg_quote($keyword, '/') . '\b/', $linked, $content, 1);
            $added[] = $link_key;
        }
    }
    
    // Add resource box at bottom
    if (!empty($added)) {
        $content .= '<div style="background:linear-gradient(135deg,rgba(212,175,55,0.08),rgba(212,175,55,0.02));border:1px solid rgba(212,175,55,0.2);border-radius:16px;padding:1.5em;margin-top:2em;">';
        $content .= '<h4 style="color:#FFD700;margin-top:0;">Recommended Tools & Resources</h4>';
        $content .= '<ul style="list-style:none;padding:0;margin:0;">';
        
        foreach ($added as $link_key) {
            $link = $links[$link_key];
            $content .= '<li style="margin-bottom:0.5em;"><a href="' . esc_url($link['url']) . '?ref=aimoneymachine" target="_blank" rel="nofollow noopener sponsored" style="color:#D4AF37;">' . $link['text'] . '</a> — <em style="color:#A89E90;">' . $link['commission'] . ' commission</em></li>';
        }
        
        $content .= '</ul>';
        $content .= '<p style="color:#6B6358;font-size:0.85em;margin-bottom:0;margin-top:1em;"><em>Disclosure: We may earn a commission if you sign up through our links at no extra cost to you.</em></p>';
        $content .= '</div>';
    }
    
    return $content;
}
add_filter('the_content', 'add_affiliate_links_to_content');

// Add affiliate disclosure to all posts
function add_affiliate_disclosure($content) {
    if (!is_single()) return $content;
    
    $disclosure = '<div style="background:rgba(212,175,55,0.05);border-left:3px solid #B8860B;padding:0.8em 1em;margin-bottom:1.5em;border-radius:0 8px 8px 0;font-size:0.85em;color:#A89E90;">';
    $disclosure .= '<strong style="color:#D4AF37;">Disclosure:</strong> This post may contain affiliate links. We earn a commission if you make a purchase through our links at no extra cost to you.';
    $disclosure .= '</div>';
    
    return $disclosure . $content;
}
add_filter('the_content', 'add_affiliate_disclosure', 5);

// Track affiliate clicks
function track_affiliate_click() {
    if (isset($_GET['ref']) && $_GET['ref'] === 'aimoneymachine') {
        $clicks = get_option('affiliate_clicks', array());
        $today = date('Y-m-d');
        if (!isset($clicks[$today])) $clicks[$today] = 0;
        $clicks[$today]++;
        update_option('affiliate_clicks', $clicks);
    }
}
add_action('init', 'track_affiliate_click');
