async (page) => {
  await page.goto('http://127.0.0.1:7777');
  await page.waitForLoadState('networkidle');
  await page.waitForFunction(() => {
    const imgs = [...document.querySelectorAll('img')].filter((candidate) => candidate.currentSrc.includes('/api/previews/') || candidate.src.includes('/api/previews/'));
    return imgs.some((img) => img.complete && img.naturalWidth > 0 && img.naturalHeight > 0);
  }, { timeout: 20000 });
  return await page.evaluate(() => ({
    action: 'open-second-desktop-tab',
    previews: [...document.querySelectorAll('img')].map((img) => ({
      alt: img.getAttribute('alt') || '',
      currentSrc: img.currentSrc || '',
      naturalWidth: img.naturalWidth,
      naturalHeight: img.naturalHeight,
    })),
  }));
}
