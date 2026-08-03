#!/usr/bin/env php
<?php
/**
 * Scan all published WP posts for raw markdown syntax.
 * Run: php /tmp/scan_markdown.php
 */
$_SERVER['SERVER_NAME'] = 'aimoneymachine.site';
define('ABSPATH', '/var/www/aimoneymachine/');
require_once ABSPATH . 'wp-load.php';

global $wpdb;
$posts = $wpdb->get_results(
    "SELECT ID, post_title, post_content FROM {$wpdb->posts} WHERE post_status='publish' AND post_type='post'"
);

echo "Scanning " . count($posts) . " published posts...\n\n";

$stats = array('heading'=>0, 'bold'=>0, 'italic'=>0, 'link'=>0, 'image'=>0,
               'codeblock'=>0, 'blockquote'=>0, 'raw_model'=>0, 'hr'=>0, 'underline_md'=>0);
$affected = array();
$total_issues = 0;

foreach ($posts as $post) {
    $content = $post->post_content;
    $issues = array();
    if (empty($content)) continue;

    // Markdown headings: # at start of line
    if (preg_match_all('/(?:^|\n)(#{1,6})\s([^\n]+)/', $content, $m)) {
        foreach ($m[0] as $match) {
            $issues[] = array('heading', 0, substr(trim($match), 0, 80));
        }
    }

    // Bold **text**
    if (preg_match_all('/(?<!<)\*\*([^*<>]+?)\*\*(?!>)/', $content, $m)) {
        foreach (array_slice($m[1], 0, 5) as $match) {
            $issues[] = array('bold', 0, substr($match, 0, 60));
        }
    }

    // Italic *text*
    if (preg_match_all('/(?<![<*])\*(?!\*)\s*([^*<>]+?)\s*\*(?![*>])/', $content, $m)) {
        foreach (array_slice($m[1], 0, 3) as $match) {
            $issues[] = array('italic', 0, substr($match, 0, 60));
        }
    }

    // Underline __text__
    if (preg_match_all('/(?<!\w)__([^_]+?)__(?!\w)/', $content, $m)) {
        foreach (array_slice($m[1], 0, 3) as $match) {
            $issues[] = array('underline_md', 0, substr($match, 0, 60));
        }
    }

    // Markdown links [text](url)
    if (preg_match_all('/\[([^\]]+)\]\((https?:\/\/[^)]+)\)/', $content, $m)) {
        foreach (array_slice($m[1], 0, 3) as $match) {
            $issues[] = array('link', 0, substr($match, 0, 60));
        }
    }

    // Markdown images ![alt](url)
    if (preg_match_all('/!\[([^\]]*)\]\((https?:\/\/[^)]+)\)/', $content, $m)) {
        foreach (array_slice($m[1], 0, 3) as $match) {
            $issues[] = array('image', 0, substr($match, 0, 60));
        }
    }

    // Code blocks ```
    $cb_count = substr_count($content, '```');
    if ($cb_count > 0) {
        $issues[] = array('codeblock', 0, "$cb_count backtick blocks");
    }

    // Blockquotes
    if (preg_match_all('/(?:^|\n)>\s([^\n]+)/', $content, $m)) {
        $issues[] = array('blockquote', 0, count($m[1]) . ' lines');
    }

    // Raw model/LLM tags
    if (stripos($content, '[Model:') !== false || stripos(substr($content, 0, 200), 'gpt-') !== false) {
        $issues[] = array('raw_model', 0, 'LLM model tag found');
    }

    // Horizontal rule ---
    if (preg_match('/(?:^|\n)---+(?:\n|$)/', $content)) {
        $issues[] = array('hr', 0, 'horizontal rule');
    }

    if (!empty($issues)) {
        $affected[] = array('id' => $post->ID, 'title' => $post->post_title, 'issues' => $issues);
        $total_issues += count($issues);
        foreach ($issues as $issue) {
            if (isset($stats[$issue[0]])) {
                $stats[$issue[0]]++;
            }
        }
    }
}

echo str_repeat('=', 70) . "\n";
echo "MARKDOWN SYNTAX SCAN RESULTS\n";
echo str_repeat('=', 70) . "\n";
echo "Total posts scanned: " . count($posts) . "\n";
echo "Posts affected: " . count($affected) . "\n";
echo "Total issue instances: $total_issues\n\n";
echo "Breakdown by type:\n";
arsort($stats);
foreach ($stats as $k => $v) {
    if ($v > 0) {
        printf("  %-15s: %d posts\n", $k, $v);
    }
}

echo "\nMost affected posts (top 20):\n";
usort($affected, function($a, $b) { return count($b['issues']) - count($a['issues']); });
foreach (array_slice($affected, 0, 20) as $a) {
    $types = array_unique(array_map(function($i) { return $i[0]; }, $a['issues']));
    printf("  [%4d] %-55s | %2d issues: %s\n", $a['id'], substr($a['title'], 0, 55), count($a['issues']), implode(', ', $types));
}

// Save full list
file_put_contents('/tmp/markdown_scan.json', json_encode($affected, JSON_PRETTY_PRINT));
echo "\nFull list saved to /tmp/markdown_scan.json\n";
