<?php
// Test Twitter API v2
$api_key = "hrHCxGaQi8ztbfaG0cs8sjfl2";
$api_secret = "A6hDigI5JbnjQQR1NvtvOGl3BdqoBPtGNrJkVPt7nDYX0MBWBU";
$access_token = "2044648665470406656-K6ZREJ7Ab7ve6zx0T9QPcjoTVf9vEg";
$access_secret = "HBJr0tARufOKDsAKTUXt14YNr06VO3KVrj45GfG70d6La";

// Twitter API v2 endpoint for posting tweets
$url = "https://api.twitter.com/2/tweets";

// Create tweet payload
$payload = json_encode(array(
    "text" => "Test post from AI Money Machine - Automated content publishing is live! #AIMoney #Automation"
));

// Generate OAuth signature for v2
$oauth_params = array(
    "oauth_consumer_key" => $api_key,
    "oauth_nonce" => md5(uniqid(mt_rand(), true)),
    "oauth_signature_method" => "HMAC-SHA1",
    "oauth_timestamp" => time(),
    "oauth_token" => $access_token,
    "oauth_version" => "1.0"
);

// For v2, we need to use the request body in the signature
$request_body_hash = base64_encode(hash("sha256", $payload, true));
$oauth_params["oauth_body_hash"] = $request_body_hash;

ksort($oauth_params);

$base_string = "POST&" . urlencode($url) . "&" . urlencode(http_build_query($oauth_params, "", "&", PHP_QUERY_RFC3986));
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
curl_setopt($ch, CURLOPT_POSTFIELDS, $payload);
curl_setopt($ch, CURLOPT_HTTPHEADER, array(
    "Authorization: " . $auth_header,
    "Content-Type: application/json"
));
curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
$response = curl_exec($ch);
$http_code = curl_getinfo($ch, CURLINFO_HTTP_CODE);
curl_close($ch);

echo "HTTP Code: " . $http_code . "\n";
echo "Response: " . $response . "\n";
