#!/usr/bin/env php
<?php
/**
 * Fix markdown syntax in all published WP posts.
 * Converts markdown to proper HTML for WordPress.
 * Run: php /tmp/fix_markdown.php [--dry-run]
 */
$_SERVER['SERVER_NAME'] = 'aimoneymachine.site';
define('ABSPATH', '/var/www/aimoneymachine/');
require_once ABSPATH . 'wp-load.php';

$dry_run = in_array('--dry-run', $argv);
$batch_size = 20; // Process in batches to avoid timeouts

global $wpdb;

// Get all affected posts
$posts = $wpdb->get_results(
    "SELECT ID, post_title, post_content FROM {$wpdb->posts} 
     WHERE post_status='publish' AND post_type='post' 
     AND (post_content LIKE '%## %' 
       OR post_content LIKE '%**%**%' 
       OR post_content LIKE '%```%'
       OR post_content LIKE '%[Model:%')"
);

echo ($dry_run ? "[DRY RUN] " : "") . "Processing " . count($posts) . " affected posts...\n\n";

$fixed = 0;
$skipped = 0;
$errors = 0;

foreach ($posts as $post) {
    $original = $post->post_content;
    $content = $original;
    $changes = array();

    // 1. Convert markdown headings to HTML
    // Only convert if NOT already inside HTML tags
    $content = preg_replace_callback('/(?:^|\n)(#{1,6})\s([^\n]+)/', function($m) use (&$changes) {
        $level = strlen($m[1]);
        $text = trim($m[2]);
        // Don't convert if it looks like it's already HTML
        if (strpos($text, '<h') === 0) return $m[0];
        $changes[] = "h$level: " . substr($text, 0, 40);
        return "\n<h$level>" . $text . "</h$level>";
    }, $content);

    // 2. Convert bold **text** to <strong>
    $content = preg_replace_callback('/(?<!<)\*\*([^*<>]+?)\*\*(?!>)/', function($m) use (&$changes) {
        $changes[] = "bold: " . substr($m[1], 0, 30);
        return '<strong>' . $m[1] . '</strong>';
    }, $content);

    // 3. Convert italic *text* to <em> (but not inside words or HTML)
    $content = preg_replace_callback('/(?<![<*])\*(?!\*)\s*([^*<>]+?)\s*\*(?![*>])/', function($m) use (&$changes) {
        $changes[] = "italic: " . substr($m[1], 0, 30);
        return '<em>' . trim($m[1]) . '</em>';
    }, $content);

    // 4. Convert underline __text__ to <u>
    $content = preg_replace_callback('/(?<!\w)__([^_]+?)__(?!\w)/', function($m) use (&$changes) {
        $changes[] = "underline: " . substr($m[1], 0, 30);
        return '<u>' . $m[1] . '</u>';
    }, $content);

    // 5. Convert markdown links [text](url) to <a>
    $content = preg_replace_callback('/\[([^\]]+)\]\((https?:\/\/[^)]+)\)/', function($m) use (&$changes) {
        $changes[] = "link: " . substr($m[1], 0, 30);
        return '<a href="' . $m[2] . '" target="_blank" rel="noopener">' . $m[1] . '</a>';
    }, $content);

    // 6. Convert markdown images ![alt](url) to <img>
    $content = preg_replace_callback('/!\[([^\]]*)\]\((https?:\/\/[^)]+)\)/', function($m) use (&$changes) {
        $changes[] = "image: " . substr($m[1], 0, 30);
        return '<img src="' . $m[2] . '" alt="' . $m[1] . '" />';
    }, $content);

    // 7. Convert code blocks ``` to <pre><code>
    $content = preg_replace_callback('/```(\w*)\n(.*?)```/s', function($m) use (&$changes) {
        $changes[] = "codeblock";
        $lang = $m[1] ? ' class="language-' . $m[1] . '"' : '';
        return '<pre><code' . $lang . '>' . htmlspecialchars($m[2]) . '</code></pre>';
    }, $content);
    // Handle unclosed code blocks (no closing ```)
    $content = preg_replace_callback('/```(\w*)\n(.*?)(?:\n\n|$)/s', function($m) use (&$changes) {
        $changes[] = "codeblock (unclosed)";
        $lang = $m[1] ? ' class="language-' . $m[1] . '"' : '';
        return '<pre><code' . $lang . '>' . htmlspecialchars($m[2]) . '</code></pre>';
    }, $content);

    // 8. Convert blockquotes > text to <blockquote>
    $content = preg_replace_callback('/(?:^|\n)>\s([^\n]+(?:\n>\s[^\n]+)*)/', function($m) use (&$changes) {
        $changes[] = "blockquote";
        $text = preg_replace('/^>\s?/m', '', $m[1]);
        return "\n<blockquote><p>" . nl2br($text) . "</p></blockquote>";
    }, $content);

    // 9. Convert horizontal rule --- to <hr>
    $content = preg_replace_callback('/(?:^|\n)---+(?:\n|$)/', function($m) use (&$changes) {
        $changes[] = "hr";
        return "\n<hr />\n";
    }, $content);

    // 10. Remove raw model tags [Model: gpt...]
    $content = preg_replace('/\[Model:[^\]]*\]/i', '', $content);
    $content = preg_replace('/\[?[Gg]pt-[^\s\]]*\]?/', '', $content);

    // 11. Clean up extra newlines from conversions
    $content = preg_replace('/\n{3,}/', "\n\n", $content);

    // Check if content actually changed
    if ($content === $original) {
        $skipped++;
        continue;
    }

    if ($dry_run) {
        echo "[{$post->ID}] {$post->post_title}\n";
        echo "  Changes: " . implode(', ', array_slice($changes, 0, 5)) . "\n";
        if (count($changes) > 5) {
            echo "  ... and " . (count($changes) - 5) . " more\n";
        }
        echo "\n";
        $fixed++;
    } else {
        // Update the post
        $result = $wpdb->update(
            $wpdb->posts,
            array('post_content' => $content),
            array('ID' => $post->ID),
            array('%s'),
            array('%d')
        );
        if ($result !== false) {
            $fixed++;
            // Clear any caches
            clean_post_cache($post->ID);
        } else {
            $errors++;
            echo "[ERROR] Failed to update post {$post->ID}: " . $wpdb->last_error . "\n";
        }
    }
}

echo str_repeat('=', 70) . "\n";
echo "SUMMARY\n";
echo str_repeat('=', 70) . "\n";
echo "Posts processed: " . count($posts) . "\n";
echo "Posts fixed: $fixed\n";
echo "Posts skipped (no changes): $skipped\n";
echo "Errors: $errors\n";
echo ($dry_run ? "\n[DRY RUN] No changes were made. Run without --dry-run to apply.\n" : "\nAll changes applied!\n");
