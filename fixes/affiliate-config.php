<?php
/**
 * Plugin Name: Affiliate Links Config
 * Description: Affiliate links configuration - functions handled by analytics-tracking.php and post-optimizer.php
 */

// Affiliate link database (used by post-optimizer.php)
function get_affiliate_links() {
    return array(
        'chatgpt' => array('url' => 'https://openai.com/chatgpt/pricing/', 'text' => 'ChatGPT Plus', 'commission' => '$20/mo'),
        'claude' => array('url' => 'https://claude.ai/pricing', 'text' => 'Claude Pro', 'commission' => '$20/mo'),
        'midjourney' => array('url' => 'https://www.midjourney.com/', 'text' => 'Midjourney', 'commission' => '20%'),
        'jasper' => array('url' => 'https://www.jasper.ai/', 'text' => 'Jasper AI', 'commission' => '30%'),
        'copyai' => array('url' => 'https://www.copy.ai/', 'text' => 'Copy.ai', 'commission' => '45%'),
        'writesonic' => array('url' => 'https://writesonic.com/', 'text' => 'Writesonic', 'commission' => '30%'),
        'grammarly' => array('url' => 'https://www.grammarly.com/', 'text' => 'Grammarly', 'commission' => '$20/sale'),
        'surfer' => array('url' => 'https://surferseo.com/', 'text' => 'Surfer SEO', 'commission' => '25%'),
        'binance' => array('url' => 'https://www.binance.com/en/register', 'text' => 'Binance', 'commission' => '50% trading fees'),
        'coinbase' => array('url' => 'https://www.coinbase.com/', 'text' => 'Coinbase', 'commission' => '$10/signup'),
        'kraken' => array('url' => 'https://www.kraken.com/', 'text' => 'Kraken', 'commission' => '20% trading fees'),
        'cryptocom' => array('url' => 'https://crypto.com/', 'text' => 'Crypto.com', 'commission' => '$25/signup'),
        'hetzner' => array('url' => 'https://www.hetzner.com/cloud', 'text' => 'Hetzner Cloud', 'commission' => '€20/signup'),
        'digitalocean' => array('url' => 'https://www.digitalocean.com/', 'text' => 'DigitalOcean', 'commission' => '$200 credit'),
        'aws' => array('url' => 'https://aws.amazon.com/free/', 'text' => 'AWS Free Tier', 'commission' => 'varies'),
        'cloudflare' => array('url' => 'https://www.cloudflare.com/', 'text' => 'Cloudflare', 'commission' => 'varies'),
        'udemy' => array('url' => 'https://www.udemy.com/', 'text' => 'Udemy', 'commission' => '15%'),
        'coursera' => array('url' => 'https://www.coursera.org/', 'text' => 'Coursera', 'commission' => '10-45%'),
        'skillshare' => array('url' => 'https://www.skillshare.com/', 'text' => 'Skillshare', 'commission' => '$7/trial'),
        'notion' => array('url' => 'https://www.notion.so/', 'text' => 'Notion', 'commission' => '$10/signup'),
        'zapier' => array('url' => 'https://zapier.com/', 'text' => 'Zapier', 'commission' => 'varies'),
        'make' => array('url' => 'https://www.make.com/', 'text' => 'Make', 'commission' => 'varies'),
    );
}
