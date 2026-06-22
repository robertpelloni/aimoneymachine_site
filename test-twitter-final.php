<?php
// Test Twitter API v2 with all new credentials
$api_key = "PlRsMkKpUGOzMOaumlWI9GtxA";
$api_secret = "CrEQnwjYAPOXKX2SlWZQILmOcqlkMbIuahTQvVuiCCGZ6I5aXF";
$access_token = "2044648665470406656-zXTs7oqOjOOdNQpM7EEoSkciSTRbBQ";
$access_secret = "dCyrB5bnc6MZS2kwweRB4buoFEr1Pbl7U2f8dwTHRNtVi";

// First verify credentials
echo "Verifying credentials...\n";
$url = "https://api.twitter.com/2/users/me";

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

echo "User Info - HTTP Code: " . $http_code . "\n";
echo "Response: " . $response . "\n\n";

if ($http_code == 200) {
    echo "Credentials are valid! Posting test tweet...\n\n";
    
    // Now try posting a tweet
    $url = "https://api.twitter.com/2/tweets";
    $payload = json_encode(array(
        "text" => "AI Money Machine is live! Automated content publishing with 190+ topics and 49 affiliate programs. #AIMoney #PassiveIncome #Automation"
    ));
    
    $oauth_params = array(
        "oauth_consumer_key" => $api_key,
        "oauth_nonce" => md5(uniqid(mt_rand(), true)),
        "oauth_signature_method" => "HMAC-SHA1",
        "oauth_timestamp" => time(),
        "oauth_token" => $access_token,
        "oauth_version" => "1.0"
    );
    
    $body_hash = base64_encode(hash("sha256", $payload, true));
    $oauth_params["oauth_body_hash"] = $body_hash;
    
    ksort($oauth_params);
    
    $base_string = "POST&" . urlencode($url) . "&" . urlencode(http_build_query($oauth_params, "", "&", PHP_QUERY_RFC3986));
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
    curl_setopt($ch, CURLOPT_POST, true);
    curl_setopt($ch, CURLOPT_POSTFIELDS, $payload);
    curl_setopt($ch, CURLOPT_HTTPHEADER, array(
        "Authorization: " . $auth_header,
        "Content-Type: application/json"
    ));
    curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
    $response = curl_exec($ch);
    $http_code = curl_getinfo($ch, CURLINFO_HTTP_CODE);
    curl_close($ch);
    
    echo "Tweet - HTTP Code: " . $http_code . "\n";
    echo "Response: " . $response . "\n";
}
