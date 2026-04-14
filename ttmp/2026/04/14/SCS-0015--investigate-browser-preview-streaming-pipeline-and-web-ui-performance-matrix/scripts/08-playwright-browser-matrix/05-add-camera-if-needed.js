async (page) => {
  const clickButtonByText = async (text, { exact = false } = {}) => {
    const result = await page.evaluate(({ text, exact }) => {
      const normalize = (s) => (s || '').replace(/\s+/g, ' ').trim();
      const candidates = [...document.querySelectorAll('button'), ...document.querySelectorAll('[role="button"]')];
      const match = candidates.find((el) => {
        const value = normalize(el.textContent);
        return exact ? value === text : value.includes(text);
      });
      if (!match) return { ok: false, candidates: candidates.map((el) => normalize(el.textContent)).filter(Boolean) };
      match.dispatchEvent(new MouseEvent('click', { bubbles: true, cancelable: true, view: window }));
      return { ok: true, clicked: normalize(match.textContent) };
    }, { text, exact });
    if (!result.ok) throw new Error(`failed to find button ${text}; candidates=${result.candidates.join(' | ')}`);
    return result.clicked;
  };

  const cameraPreviewReady = async () => page.evaluate(() => {
    const img = [...document.querySelectorAll('img')].find((candidate) => {
      const alt = candidate.getAttribute('alt') || '';
      return alt.includes('laptop-camera');
    });
    return !!img && img.complete && img.naturalWidth > 0 && img.naturalHeight > 0;
  });

  if (await cameraPreviewReady()) {
    return { action: 'add-camera-if-needed', skipped: true, reason: 'camera_already_present' };
  }

  await clickButtonByText('Add Source');
  await page.waitForTimeout(200);
  await clickButtonByText('◉ Camera').catch(async () => clickButtonByText('Camera', { exact: true }));
  await page.waitForTimeout(200);
  await clickButtonByText('Laptop Camera:');
  await page.waitForFunction(() => {
    const img = [...document.querySelectorAll('img')].find((candidate) => {
      const alt = candidate.getAttribute('alt') || '';
      return alt.includes('laptop-camera');
    });
    return !!img && img.complete && img.naturalWidth > 0 && img.naturalHeight > 0;
  }, { timeout: 15000 });
  return { action: 'add-camera-if-needed', added: true };
}
