<?php
/**
 * Plugin Name: Twitter Auto-Post
 * Description: Automatically posts to Twitter when new content is published
 */

function post_to_twitter($post_id) {
    // Only post for published posts
    if (get_post_status($post_id) !== 'publish') return;
    if (wp_is_post_revision($post_id)) return;
    
    // Get Twitter API credentials
    $api_key = get_option('twitter_api_key');
    $api_secret = get_option('twitter_api_secret');
    $access_token = get_option('twitter_access_token');
    $access_secret = get_option('twitter_access_secret');
    
    if (empty($api_key) || empty($access_token)) return;
    
    // Get post data
    $title = get_the_title($post_id);
    $url = get_permalink($post_id);
    $excerpt = wp_trim_words(get_post_field('post_content', $post_id), 20, '...');
    
    // Create tweet with hashtags
    $tweet = "$title\n\n$excerpt\n\n$url\n\n#AIMoney #PassiveIncome #AI #Automation #SideHustle";
    
    // Truncate to 280 characters
    if (strlen($tweet) > 280) {
        $tweet = substr($tweet, 0, 277) . '...';
    }
    
    // Post to Twitter using OAuth 1.0a
    $result = twitter_api_post($tweet, $api_key, $api_secret, $access_token, $access_secret);
    
    if ($result) {
        update_post_meta($post_id, '_twitter_posted', current_time('mysql'));
        update_post_meta($post_id, '_twitter_tweet_id', $result);
    }
}
add_action('publish_post', 'post_to_twitter');

function twitter_api_post($tweet, $api_key, $api_secret, $access_token, $access_secret) {
    $url = 'https://api.twitter.com/1.1/statuses/update.json';
    
    $params = array(
        'status' => $tweet
    );
    
    // Generate OAuth signature
    $oauth_params = array(
        'oauth_consumer_key' => $api_key,
        'oauth_nonce' => md5(uniqid(mt_rand(), true)),
        'oauth_signature_method' => 'HMAC-SHA1',
        'oauth_timestamp' => time(),
        'oauth_token' => $access_token,
        'oauth_version' => '1.0'
    );
    
    $all_params = array_merge($oauth_params, $params);
    ksort($all_params);
    
    $base_string = 'POST&' . urlencode($url) . '&' . urlencode(http_build_query($all_params, '', '&', PHP_QUERY_RFC3986));
    $signing_key = urlencode($api_secret) . '&' . urlencode($access_secret);
    $signature = base64_encode(hash_hmac('sha1', $base_string, $signing_key, true));
    
    $oauth_params['oauth_signature'] = $signature;
    
    // Build Authorization header
    $auth_header = 'OAuth ';
    $auth_parts = array();
    foreach ($oauth_params as $key => $value) {
        $auth_parts[] = $key . '="' . urlencode($value) . '"';
    }
    $auth_header .= implode(', ', $auth_parts);
    
    // Make request
    $response = wp_remote_post($url, array(
        'headers' => array(
            'Authorization' => $auth_header
        ),
        'body' => $params
    ));
    
    if (is_wp_error($response)) {
        error_log('Twitter API error: ' . $response->get_error_message());
        return false;
    }
    
    $body = json_decode(wp_remote_retrieve_body($response), true);
    
    if (isset($body['id_str'])) {
        return $body['id_str'];
    }
    
    error_log('Twitter API response: ' . print_r($body, true));
    return false;
}
