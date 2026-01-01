#!/usr/bin/env python3
"""Test Open Notebook UI and capture screenshots"""

from playwright.sync_api import sync_playwright

with sync_playwright() as p:
    browser = p.chromium.launch(headless=True)
    page = browser.new_page(viewport={'width': 1400, 'height': 900})

    # Navigate to the app
    page.goto('http://localhost:8080/')
    page.wait_for_load_state('networkidle')

    # Capture full page screenshot
    page.screenshot(path='/tmp/open-notebook-ui.png', full_page=True)

    # Check the CSS variables are applied
    bg_primary = page.evaluate('getComputedStyle(document.documentElement).getPropertyValue("--bg-primary")')
    accent_primary = page.evaluate('getComputedStyle(document.documentElement).getPropertyValue("--accent-primary")')

    print(f"Background Primary: {bg_primary}")
    print(f"Accent Primary: {accent_primary}")

    # Get page title
    title = page.title()
    print(f"Page Title: {title}")

    # Check for chat elements
    chat_input = page.locator('.chat-input').count()
    send_btn = page.locator('.btn-send').count()
    chat_messages = page.locator('.chat-messages').count()

    print(f"Chat Input found: {chat_input > 0}")
    print(f"Send Button found: {send_btn > 0}")
    print(f"Chat Messages container found: {chat_messages > 0}")

    # Check styling of chat input
    if chat_input > 0:
        input_styles = page.locator('.chat-input').evaluate('''el => {
            const styles = window.getComputedStyle(el);
            return {
                borderRadius: styles.borderRadius,
                backgroundColor: styles.backgroundColor,
                borderColor: styles.borderColor
            };
        }''')
        print(f"Chat Input Styles: {input_styles}")

    # Check button gradient
    if send_btn > 0:
        btn_styles = page.locator('.btn-send').evaluate('''el => {
            const styles = window.getComputedStyle(el);
            return {
                background: styles.background,
                borderRadius: styles.borderRadius
            };
        }''')
        print(f"Send Button Styles: {btn_styles}")

    # Check for notebook elements
    notebook_list = page.locator('.notebook-list').count()
    print(f"Notebook List found: {notebook_list > 0}")

    # Check for sources grid
    sources_grid = page.locator('.sources-grid').count()
    print(f"Sources Grid found: {sources_grid > 0}")

    browser.close()

    print("\nScreenshot saved to: /tmp/open-notebook-ui.png")
