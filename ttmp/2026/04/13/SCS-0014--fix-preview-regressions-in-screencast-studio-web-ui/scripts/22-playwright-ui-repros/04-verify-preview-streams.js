async (page) => {
  await page.goto('http://127.0.0.1:7777');
  await page.waitForLoadState('networkidle');

  await page.waitForFunction(() => document.querySelectorAll('img').length > 0, { timeout: 10000 });
  await page.waitForTimeout(1000);

  return await page.evaluate(() => {
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
      action: 'verify-preview-streams',
      sourceCountText: text.match(/Sources \((\d+)\)/)?.[0] || null,
      previews,
      livePreviews: previews.filter((preview) => preview.complete && preview.naturalWidth > 0 && preview.naturalHeight > 0),
    };
  });
}
