<?php
/**
 * Plugin Name: Affiliate Links Inserter
 * Description: Adds contextual affiliate links to post content
 */

// Check if function already exists before declaring
if (!function_exists('insert_affiliate_links')) {
    function insert_affiliate_links($content) {
        if (!is_single()) return $content;
        
        // Only add links if not already linked
        $links = array(
            // AI Tools - use redirect links for tracking
            '<a href="/go/chatgpt/" rel="nofollow noopener sponsored"' => 'ChatGPT',
            '<a href="/go/claude/" rel="nofollow noopener sponsored"' => 'Claude',
            '<a href="/go/midjourney/" rel="nofollow noopener sponsored"' => 'Midjourney',
            '<a href="/go/jasper/" rel="nofollow noopener sponsored"' => 'Jasper AI',
            '<a href="/go/grammarly/" rel="nofollow noopener sponsored"' => 'Grammarly',
            '<a href="/go/surfer/" rel="nofollow noopener sponsored"' => 'Surfer SEO',
            
            // Crypto
            '<a href="/go/binance/" rel="nofollow noopener sponsored"' => 'Binance',
            '<a href="/go/coinbase/" rel="nofollow noopener sponsored"' => 'Coinbase',
            '<a href="/go/kraken/" rel="nofollow noopener sponsored"' => 'Kraken',
            
            // Hosting
            '<a href="/go/hetzner/" rel="nofollow noopener sponsored"' => 'Hetzner',
            '<a href="/go/digitalocean/" rel="nofollow noopener sponsored"' => 'DigitalOcean',
            
            // Courses
            '<a href="/go/udemy/" rel="nofollow noopener sponsored"' => 'Udemy',
            '<a href="/go/coursera/" rel="nofollow noopener sponsored"' => 'Coursera',
            '<a href="/go/skillshare/" rel="nofollow noopener sponsored"' => 'Skillshare',
            
            // Productivity
            '<a href="/go/notion/" rel="nofollow noopener sponsored"' => 'Notion',
            '<a href="/go/zapier/" rel="nofollow noopener sponsored"' => 'Zapier',
        );
        
        $added = array();
        
        foreach ($links as $html => $keyword) {
            // Skip if already linked
            if (in_array($keyword, $added)) continue;
            if (stripos($content, 'href=') !== false && stripos($content, $keyword) !== false) {
                // Check if keyword is already inside an <a> tag
                $pattern = '/<a[^>]*>[^<]*' . preg_quote($keyword, '/') . '[^<]*<\/a>/i';
                if (preg_match($pattern, $content)) continue;
            }
            
            // Add link to first occurrence
            if (stripos($content, $keyword) !== false) {
                $linked = $html . '>' . $keyword . '</a>';
                $content = preg_replace('/(?<![">])' . preg_quote($keyword, '/') . '(?!<\/a>)/', $linked, $content, 1);
                $added[] = $keyword;
            }
        }
        
        return $content;
    }
    add_filter('the_content', 'insert_affiliate_links', 15);
}

// Create redirect endpoints
if (!function_exists('create_affiliate_redirects')) {
    function create_affiliate_redirects() {
        $redirects = array(
            'chatgpt' => 'https://openai.com/chatgpt/?ref=aimoneymachine',
            'claude' => 'https://claude.ai/?ref=aimoneymachine',
            'midjourney' => 'https://www.midjourney.com/?ref=aimoneymachine',
            'jasper' => 'https://www.jasper.ai/?ref=aimoneymachine',
            'grammarly' => 'https://www.grammarly.com/?ref=aimoneymachine',
            'surfer' => 'https://surferseo.com/?ref=aimoneymachine',
            'binance' => 'https://www.binance.com/en/register?ref=aimoneymachine',
            'coinbase' => 'https://www.coinbase.com/?ref=aimoneymachine',
            'kraken' => 'https://www.kraken.com/?ref=aimoneymachine',
            'hetzner' => 'https://www.hetzner.com/cloud?ref=aimoneymachine',
            'digitalocean' => 'https://www.digitalocean.com/?ref=aimoneymachine',
            'udemy' => 'https://www.udemy.com/?ref=aimoneymachine',
            'coursera' => 'https://www.coursera.org/?ref=aimoneymachine',
            'skillshare' => 'https://www.skillshare.com/?ref=aimoneymachine',
            'notion' => 'https://www.notion.so/?ref=aimoneymachine',
            'zapier' => 'https://zapier.com/?ref=aimoneymachine',
        );
        
        // Store redirects in options
        if (!get_option('affiliate_redirects')) {
            update_option('affiliate_redirects', $redirects);
        }
    }
    add_action('init', 'create_affiliate_redirects');
}

// Handle redirect requests
if (!function_exists('handle_affiliate_redirect')) {
    function handle_affiliate_redirect() {
        if (isset($_SERVER['REQUEST_URI']) && preg_match('#^/go/([a-z]+)/?$#', $_SERVER['REQUEST_URI'], $matches)) {
            $redirects = get_option('affiliate_redirects', array());
            $slug = $matches[1];
            
            if (isset($redirects[$slug])) {
                // Track the click
                $clicks = get_option('affiliate_clicks_detail', array());
                if (!isset($clicks[$slug])) $clicks[$slug] = 0;
                $clicks[$slug]++;
                update_option('affiliate_clicks_detail', $clicks);
                
                wp_redirect($redirects[$slug]);
                exit;
            }
        }
    }
    add_action('init', 'handle_affiliate_redirect');
}
