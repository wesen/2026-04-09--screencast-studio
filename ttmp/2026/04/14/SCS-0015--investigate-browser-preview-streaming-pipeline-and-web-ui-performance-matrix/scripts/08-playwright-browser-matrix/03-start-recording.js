async (page) => {
  const clickButtonByText = async (text, { exact = false } = {}) => {
    const result = await page.evaluate(({ text, exact }) => {
      const normalize = (s) => (s || '').replace(/\s+/g, ' ').trim();
      const candidates = [...document.querySelectorAll('button'), ...document.querySelectorAll('[role="button"]')];
      const match = candidates.find((el) => {
        const value = normalize(el.textContent);
        return exact ? value === text : value.includes(text);
      });
      if (!match) {
        return { ok: false, candidates: candidates.map((el) => normalize(el.textContent)).filter(Boolean) };
      }
      match.dispatchEvent(new MouseEvent('click', { bubbles: true, cancelable: true, view: window }));
      return { ok: true, clicked: normalize(match.textContent) };
    }, { text, exact });
    if (!result.ok) throw new Error(`failed to find button ${text}; candidates=${result.candidates.join(' | ')}`);
    return result.clicked;
  };

  const initialText = await page.evaluate(() => document.body.innerText || '');
  if (initialText.includes('Status: ● Recording')) {
    return { action: 'start-recording', skipped: true, reason: 'recording_already_active' };
  }

  const clicked = await clickButtonByText('Rec');
  await page.waitForFunction(() => {
    const text = document.body.innerText || '';
    return text.includes('Status: ● Recording') && text.includes('◼ Stop');
  }, { timeout: 15000 });
  return { action: 'start-recording', clicked };
}
