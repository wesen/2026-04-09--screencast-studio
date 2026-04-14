async (page) => {
  const clickButtonByText = async (text) => {
    const result = await page.evaluate((text) => {
      const normalize = (s) => (s || '').replace(/\s+/g, ' ').trim();
      const candidates = [...document.querySelectorAll('button'), ...document.querySelectorAll('[role="button"]')];
      const match = candidates.find((el) => normalize(el.textContent).includes(text));
      if (!match) return { ok: false, candidates: candidates.map((el) => normalize(el.textContent)).filter(Boolean) };
      match.dispatchEvent(new MouseEvent('click', { bubbles: true, cancelable: true, view: window }));
      return { ok: true, clicked: normalize(match.textContent) };
    }, text);
    if (!result.ok) throw new Error(`failed to find button ${text}; candidates=${result.candidates.join(' | ')}`);
    return result.clicked;
  };

  const initialText = await page.evaluate(() => document.body.innerText || '');
  if (!initialText.includes('Status: ● Recording')) {
    return { action: 'stop-recording', skipped: true, reason: 'recording_not_active' };
  }

  const clicked = await clickButtonByText('Stop');
  await page.waitForFunction(() => {
    const text = document.body.innerText || '';
    return text.includes('Status: ◻ Ready') || text.includes('● Rec');
  }, { timeout: 20000 });
  return { action: 'stop-recording', clicked };
}
