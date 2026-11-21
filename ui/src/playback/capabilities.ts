let hevc: '0' | '1' | undefined;

export function browserHEVC() {
  if (hevc === undefined) {
    const video = document.createElement('video');
    // Main and Main 10 at level 5 cover common 4K sources at up to 30 fps.
    const codecs = ['hvc1.1.6.L150.B0', 'hvc1.2.4.L150.B0'];
    const supported = codecs.some((codec) => video.canPlayType(`video/mp4; codecs="${codec}, mp4a.40.2"`));
    hevc = supported ? '1' : '0';
  }
  return hevc;
}
