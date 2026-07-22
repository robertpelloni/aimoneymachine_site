<?php
/**
 * Plugin Name: Impact Verification
 * Description: Adds Impact site verification meta tag
 */
function impact_verification_tag() {
    echo '<meta name="impact-site-verification" value="edcf902e-3dac-43d2-8097-a093e792c4ba">';
}
add_action('wp_head', 'impact_verification_tag');
