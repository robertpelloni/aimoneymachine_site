<?php
$_SERVER['SERVER_NAME'] = 'aimoneymachine.site';
define('ABSPATH', '/var/www/aimoneymachine/');
require_once ABSPATH . 'wp-load.php';
global $wpdb;

// Sample posts to check what markdown remains
$ids = [3772, 5249, 2265, 3750];
foreach ($ids as $id) {
    $post = $wpdb->get_row($wpdb->prepare("SELECT post_title, post_content FROM {$wpdb->posts} WHERE ID=%d", $id));
    if (!$post) continue;
    echo "=== [{$id}] {$post->post_title} ===\n";
    $lines = explode("\n", $post->post_content);
    $count = 0;
    foreach ($lines as $i => $line) {
        if (preg_match('/(?:^|\s)(?:\*\*[^*]+\*\*|\*[^*]+\*|#{1,6}\s|__[^_]+__|```|^> )/', $line)) {
            $display = substr(trim($line), 0, 120);
            echo "  L" . ($i+1) . ": " . $display . "\n";
            $count++;
            if ($count >= 8) { echo "  ...\n"; break; }
        }
    }
    echo "\n";
}
