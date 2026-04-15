async (page) => {
  await page.goto('http://127.0.0.1:7777');
  await page.waitForLoadState('networkidle');
  await page.waitForFunction(() => {
    const img = [...document.querySelectorAll('img')].find((candidate) => {
      const alt = candidate.getAttribute('alt') || '';
      return alt.includes('Preview of desktop-1') || alt.includes('Preview of') || candidate.src.includes('/api/previews/');
    });
    return !!img && img.complete && img.naturalWidth > 0 && img.naturalHeight > 0;
  }, { timeout: 20000 });
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
      action: 'open-studio-and-wait-desktop',
      recording: text.includes('Status: ● Recording'),
      sourceCountText: text.match(/Sources \((\d+)\)/)?.[0] || null,
      previews,
    };
  });
}
