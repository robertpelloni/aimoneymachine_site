<?php
/**
 * AI Money Machine — Minimal Effects
 */

function aimm_inject_luxury_effects() {
    if (is_admin()) return;
    ?>
    <!-- Stealth links only -->
    <div style="text-align:center;padding:0.8em 0;font-size:0.6rem;color:#4A453D;background:#060709;border-top:1px solid rgba(184,134,11,0.1);">
        <a href="https://robertpelloni.com" style="color:#6B6358;text-decoration:none;margin:0 0.5rem;">robertpelloni.com</a>
        <span style="color:#3A3530;">|</span>
        <a href="https://bobsgame.com" style="color:#6B6358;text-decoration:none;margin:0 0.5rem;">bobsgame.com</a>
        <span style="color:#3A3530;">|</span>
        <a href="https://tormentnexus.site" style="color:#6B6358;text-decoration:none;margin:0 0.5rem;">tormentnexus.site</a>
        <span style="color:#3A3530;">|</span>
        <a href="https://hypernexus.site" style="color:#6B6358;text-decoration:none;margin:0 0.5rem;">hypernexus.site</a>
    </div>
    <?php
}
add_action("wp_footer", "aimm_inject_luxury_effects");
