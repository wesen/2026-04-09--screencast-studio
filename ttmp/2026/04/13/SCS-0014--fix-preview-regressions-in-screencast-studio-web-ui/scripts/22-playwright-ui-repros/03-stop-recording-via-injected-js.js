async (page) => {
  await page.goto('http://127.0.0.1:7777');
  await page.waitForLoadState('networkidle');

  const clickButtonByText = async (text, { exact = false } = {}) => {
    const result = await page.evaluate(({ text, exact }) => {
      const normalize = (s) => (s || '').replace(/\s+/g, ' ').trim();
      const candidates = [
        ...document.querySelectorAll('button'),
        ...document.querySelectorAll('[role="button"]'),
      ];
      const match = candidates.find((el) => {
        const value = normalize(el.textContent);
        return exact ? value === text : value.includes(text);
      });
      if (!match) {
        return {
          ok: false,
          text,
          candidates: candidates.map((el) => normalize(el.textContent)).filter(Boolean),
        };
      }
      match.dispatchEvent(new MouseEvent('click', { bubbles: true, cancelable: true, view: window }));
      return { ok: true, clicked: normalize(match.textContent) };
    }, { text, exact });
    if (!result.ok) {
      throw new Error(`failed to find button for ${text}; candidates=${result.candidates.join(' | ')}`);
    }
    return result.clicked;
  };

  const wasRecording = await page.evaluate(() => (document.body.innerText || '').includes('Status: ● Recording'));
  if (!wasRecording) {
    return {
      action: 'stop-recording-via-injected-js',
      skipped: true,
      reason: 'recording_not_active',
    };
  }

  const clicked = await clickButtonByText('Stop');

  await page.waitForFunction(() => {
    const text = document.body.innerText || '';
    return text.includes('Status: ◻ Ready') && text.includes('● Rec');
  }, { timeout: 20000 });

  const after = await page.evaluate(() => ({
    text: document.body.innerText,
    stopButtonVisible: [...document.querySelectorAll('button')].some((button) => (button.textContent || '').includes('Stop')),
    recButtonVisible: [...document.querySelectorAll('button')].some((button) => (button.textContent || '').includes('Rec')),
  }));

  return {
    action: 'stop-recording-via-injected-js',
    clicked,
    readyStatusVisible: after.text.includes('Status: ◻ Ready'),
    recButtonVisible: after.recButtonVisible,
    stopButtonVisible: after.stopButtonVisible,
  };
}
