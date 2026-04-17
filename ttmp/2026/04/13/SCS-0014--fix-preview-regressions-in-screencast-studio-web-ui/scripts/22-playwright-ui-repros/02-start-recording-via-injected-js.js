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

  const before = await page.evaluate(() => ({
    text: document.body.innerText,
    timer: [...document.querySelectorAll('*')].map((el) => (el.textContent || '').trim()).find((value) => /^\d\d:\d\d:\d\d$/.test(value)) || null,
  }));

  if (before.text.includes('Status: ● Recording')) {
    return {
      action: 'start-recording-via-injected-js',
      skipped: true,
      reason: 'recording_already_active',
      beforeTimer: before.timer,
    };
  }

  const clicked = await clickButtonByText('Rec');

  await page.waitForFunction(() => {
    const text = document.body.innerText || '';
    return text.includes('Status: ● Recording') && text.includes('◼ Stop');
  }, { timeout: 15000 });

  const after = await page.evaluate(() => ({
    text: document.body.innerText,
    timer: [...document.querySelectorAll('*')].map((el) => (el.textContent || '').trim()).find((value) => /^\d\d:\d\d:\d\d$/.test(value)) || null,
    recordingBadgeVisible: (document.body.innerText || '').includes('● REC'),
    stopButtonVisible: [...document.querySelectorAll('button')].some((button) => (button.textContent || '').includes('Stop')),
  }));

  return {
    action: 'start-recording-via-injected-js',
    clicked,
    beforeTimer: before.timer,
    afterTimer: after.timer,
    recordingBadgeVisible: after.recordingBadgeVisible,
    stopButtonVisible: after.stopButtonVisible,
    recordingStatusVisible: after.text.includes('Status: ● Recording'),
  };
}
