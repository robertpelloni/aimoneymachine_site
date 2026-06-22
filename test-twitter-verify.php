<?php
// Test Twitter API credentials
$api_key = "hrHCxGaQi8ztbfaG0cs8sjfl2";
$api_secret = "A6hDigI5JbnjQQR1NvtvOGl3BdqoBPtGNrJkVPt7nDYX0MBWBU";
$access_token = "2044648665470406656-K6ZREJ7Ab7ve6zx0T9QPcjoTVf9vEg";
$access_secret = "HBJr0tARufOKDsAKTUXt14YNr06VO3KVrj45GfG70d6La";

// First verify credentials
$url = "https://api.twitter.com/1.1/account/verify_credentials.json";

$oauth_params = array(
    "oauth_consumer_key" => $api_key,
    "oauth_nonce" => md5(uniqid(mt_rand(), true)),
    "oauth_signature_method" => "HMAC-SHA1",
    "oauth_timestamp" => time(),
    "oauth_token" => $access_token,
    "oauth_version" => "1.0"
);

ksort($oauth_params);

$base_string = "GET&" . urlencode($url) . "&" . urlencode(http_build_query($oauth_params, "", "&", PHP_QUERY_RFC3986));
$signing_key = urlencode($api_secret) . "&" . urlencode($access_secret);
$signature = base64_encode(hash_hmac("sha1", $base_string, $signing_key, true));

$oauth_params["oauth_signature"] = $signature;

$auth_header = "OAuth ";
$auth_parts = array();
foreach ($oauth_params as $key => $value) {
    $auth_parts[] = $key . "=\"" . urlencode($value) . "\"";
}
$auth_header .= implode(", ", $auth_parts);

$ch = curl_init();
curl_setopt($ch, CURLOPT_URL, $url);
curl_setopt($ch, CURLOPT_HTTPHEADER, array("Authorization: " . $auth_header));
curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
$response = curl_exec($ch);
$http_code = curl_getinfo($ch, CURLINFO_HTTP_CODE);
curl_close($ch);

echo "HTTP Code: " . $http_code . "\n";
echo "Response: " . $response . "\n";

// If credentials are valid, try posting a tweet
if ($http_code == 200) {
    echo "\nCredentials are valid! Posting test tweet...\n";
    
    $tweet_url = "https://api.twitter.com/1.1/statuses/update.json";
    $tweet_params = array("status" => "Test post from AI Money Machine - Automated content publishing is live! #AIMoney #Automation");
    
    $tweet_oauth_params = array(
        "oauth_consumer_key" => $api_key,
        "oauth_nonce" => md5(uniqid(mt_rand(), true)),
        "oauth_signature_method" => "HMAC-SHA1",
        "oauth_timestamp" => time(),
        "oauth_token" => $access_token,
        "oauth_version" => "1.0"
    );
    
    $all_params = array_merge($tweet_oauth_params, $tweet_params);
    ksort($all_params);
    
    $tweet_base_string = "POST&" . urlencode($tweet_url) . "&" . urlencode(http_build_query($all_params, "", "&", PHP_QUERY_RFC3986));
    $tweet_signature = base64_encode(hash_hmac("sha1", $tweet_base_string, $signing_key, true));
    
    $tweet_oauth_params["oauth_signature"] = $tweet_signature;
    
    $tweet_auth_header = "OAuth ";
    $tweet_auth_parts = array();
    foreach ($tweet_oauth_params as $key => $value) {
        $tweet_auth_parts[] = $key . "=\"" . urlencode($value) . "\"";
    }
    $tweet_auth_header .= implode(", ", $tweet_auth_parts);
    
    $ch = curl_init();
    curl_setopt($ch, CURLOPT_URL, $tweet_url);
    curl_setopt($ch, CURLOPT_POST, true);
    curl_setopt($ch, CURLOPT_POSTFIELDS, $tweet_params);
    curl_setopt($ch, CURLOPT_HTTPHEADER, array("Authorization: " . $tweet_auth_header));
    curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
    $tweet_response = curl_exec($ch);
    $tweet_http_code = curl_getinfo($ch, CURLINFO_HTTP_CODE);
    curl_close($ch);
    
    echo "Tweet HTTP Code: " . $tweet_http_code . "\n";
    echo "Tweet Response: " . $tweet_response . "\n";
}
