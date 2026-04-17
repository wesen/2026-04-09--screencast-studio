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

  const cameraPreviewState = async () => page.evaluate(() => {
    const img = [...document.querySelectorAll('img')].find((candidate) => {
      const alt = candidate.getAttribute('alt') || '';
      return alt.includes('Preview of laptop-camera');
    });
    if (!img) {
      return null;
    }
    return {
      alt: img.getAttribute('alt') || '',
      src: img.getAttribute('src') || '',
      currentSrc: img.currentSrc || '',
      complete: img.complete,
      naturalWidth: img.naturalWidth,
      naturalHeight: img.naturalHeight,
    };
  });

  const before = await page.evaluate(() => ({
    sourceText: document.body.innerText,
    previewImages: [...document.querySelectorAll('img')].map((img) => ({
      alt: img.getAttribute('alt') || '',
      src: img.getAttribute('src') || '',
      complete: img.complete,
      naturalWidth: img.naturalWidth,
      naturalHeight: img.naturalHeight,
    })),
  }));

  let addSourceClicked = null;
  let cameraKindClicked = null;
  let deviceClicked = null;

  let cameraState = await cameraPreviewState();
  if (!cameraState) {
    addSourceClicked = await clickButtonByText('Add Source');
    await page.waitForTimeout(200);
    cameraKindClicked = await clickButtonByText('◉ Camera').catch(async () => clickButtonByText('Camera', { exact: true }));
    await page.waitForTimeout(200);
    deviceClicked = await clickButtonByText('Laptop Camera:');

    await page.waitForFunction(() => {
      const img = [...document.querySelectorAll('img')].find((candidate) => {
        const alt = candidate.getAttribute('alt') || '';
        return alt.includes('Preview of laptop-camera');
      });
      return !!img;
    }, { timeout: 10000 });

    await page.waitForFunction(() => {
      const img = [...document.querySelectorAll('img')].find((candidate) => {
        const alt = candidate.getAttribute('alt') || '';
        return alt.includes('Preview of laptop-camera');
      });
      return !!img && img.complete && img.naturalWidth > 0 && img.naturalHeight > 0;
    }, { timeout: 15000 });

    cameraState = await cameraPreviewState();
  }

  const after = await page.evaluate(() => ({
    sourceText: document.body.innerText,
    previewImages: [...document.querySelectorAll('img')].map((img) => ({
      alt: img.getAttribute('alt') || '',
      src: img.getAttribute('src') || '',
      currentSrc: img.currentSrc || '',
      complete: img.complete,
      naturalWidth: img.naturalWidth,
      naturalHeight: img.naturalHeight,
    })),
  }));

  return {
    action: 'add-camera-via-injected-js',
    addSourceClicked,
    cameraKindClicked,
    deviceClicked,
    beforeSourceCountText: before.sourceText.match(/Sources \((\d+)\)/)?.[0] || null,
    afterSourceCountText: after.sourceText.match(/Sources \((\d+)\)/)?.[0] || null,
    cameraPreview: cameraState,
    previewImages: after.previewImages,
  };
}
