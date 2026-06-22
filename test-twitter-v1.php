<?php
// Test Twitter API v1.1 with new credentials (might still work)
$api_key = "hrHCxGaQi8ztbfaG0cs8sjfl2";
$api_secret = "A6hDigI5JbnjQQR1NvtvOGl3BdqoBPtGNrJkVPt7nDYX0MBWBU";
$access_token = "2044648665470406656-Xi00tuystgpXF37IYpgeE8u4NeJFzR";
$access_secret = "eTqo69Lnq1e4B4LQxeykBdmd3WgCIn3gQMw9qdqrUtZLl";

// Try v1.1 endpoint
$url = "https://api.twitter.com/1.1/statuses/update.json";

$tweet = "AI Money Machine is live! Automated content publishing with 190+ topics and 49 affiliate programs. #AIMoney #PassiveIncome #Automation";

$params = array("status" => $tweet);

$oauth_params = array(
    "oauth_consumer_key" => $api_key,
    "oauth_nonce" => md5(uniqid(mt_rand(), true)),
    "oauth_signature_method" => "HMAC-SHA1",
    "oauth_timestamp" => time(),
    "oauth_token" => $access_token,
    "oauth_version" => "1.0"
);

$all_params = array_merge($oauth_params, $params);
ksort($all_params);

$base_string = "POST&" . urlencode($url) . "&" . urlencode(http_build_query($all_params, "", "&", PHP_QUERY_RFC3986));
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
curl_setopt($ch, CURLOPT_POST, true);
curl_setopt($ch, CURLOPT_POSTFIELDS, $params);
curl_setopt($ch, CURLOPT_HTTPHEADER, array("Authorization: " . $auth_header));
curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
$response = curl_exec($ch);
$http_code = curl_getinfo($ch, CURLINFO_HTTP_CODE);
curl_close($ch);

echo "HTTP Code: " . $http_code . "\n";
echo "Response: " . $response . "\n";
