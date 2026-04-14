async (page) => {
  return await page.evaluate(() => {
    const text = document.body.innerText || '';
    return {
      action: 'capture-browser-preview-state',
      recording: text.includes('Status: ● Recording'),
      sourceCountText: text.match(/Sources \((\d+)\)/)?.[0] || null,
      previews: [...document.querySelectorAll('img')].map((img) => ({
        alt: img.getAttribute('alt') || '',
        src: img.getAttribute('src') || '',
        currentSrc: img.currentSrc || '',
        complete: img.complete,
        naturalWidth: img.naturalWidth,
        naturalHeight: img.naturalHeight,
      })),
    };
  });
}
