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

  const cameraPreviewReady = async () => page.evaluate(() => {
    const img = [...document.querySelectorAll('img')].find((candidate) => {
      const alt = candidate.getAttribute('alt') || '';
      return alt.includes('Preview of laptop-camera');
    });
    return !!img && img.complete && img.naturalWidth > 0 && img.naturalHeight > 0;
  });

  const initialText = await page.evaluate(() => document.body.innerText || '');
  if (initialText.includes('Status: ● Recording')) {
    return {
      action: 'add-camera-and-start-recording-smoke',
      skipped: true,
      reason: 'recording_already_active',
    };
  }

  if (!(await cameraPreviewReady())) {
    await clickButtonByText('Add Source');
    await page.waitForTimeout(200);
    await clickButtonByText('◉ Camera').catch(async () => clickButtonByText('Camera', { exact: true }));
    await page.waitForTimeout(200);
    await clickButtonByText('Laptop Camera:');
    await page.waitForFunction(() => {
      const img = [...document.querySelectorAll('img')].find((candidate) => {
        const alt = candidate.getAttribute('alt') || '';
        return alt.includes('Preview of laptop-camera');
      });
      return !!img && img.complete && img.naturalWidth > 0 && img.naturalHeight > 0;
    }, { timeout: 15000 });
  }

  const startClicked = await clickButtonByText('Rec');
  await page.waitForFunction(() => {
    const text = document.body.innerText || '';
    return text.includes('Status: ● Recording') && text.includes('◼ Stop');
  }, { timeout: 15000 });

  return await page.evaluate((startClicked) => {
    const text = document.body.innerText || '';
    const previews = [...document.querySelectorAll('img')].map((img) => ({
      alt: img.getAttribute('alt') || '',
      src: img.getAttribute('src') || '',
      currentSrc: img.currentSrc || '',
      complete: img.complete,
      naturalWidth: img.naturalWidth,
      naturalHeight: img.naturalHeight,
    }));
    return {
      action: 'add-camera-and-start-recording-smoke',
      startClicked,
      sourceCountText: text.match(/Sources \((\d+)\)/)?.[0] || null,
      recordingStatusVisible: text.includes('Status: ● Recording'),
      recordingBadgeVisible: text.includes('● REC'),
      previews,
    };
  }, startClicked);
}
