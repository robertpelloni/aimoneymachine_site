<?php
/**
 * Plugin Name: Affiliate Links Manager
 * Description: Adds affiliate disclosure to posts
 */

// Add affiliate disclosure to all posts
function add_affiliate_disclosure($content) {
    if (!is_single()) return $content;
    
    $disclosure = '<div style="background:rgba(212,175,55,0.05);border-left:3px solid #B8860B;padding:0.8em 1em;margin-bottom:1.5em;border-radius:0 8px 8px 0;font-size:0.85em;color:#A89E90;">';
    $disclosure .= '<strong style="color:#D4AF37;">Disclosure:</strong> This post may contain affiliate links. We earn a commission if you make a purchase through our links at no extra cost to you.';
    $disclosure .= '</div>';
    
    return $disclosure . $content;
}
add_filter('the_content', 'add_affiliate_disclosure', 5);

// Track affiliate clicks - already handled by analytics-tracking.php
