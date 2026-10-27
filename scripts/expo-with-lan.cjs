const { withAndroidManifest, withInfoPlist } = require('expo/config-plugins');

module.exports = function withLAN(config) {
  config = withAndroidManifest(config, (value) => {
    const application = value.modResults.manifest.application[0];
    application.$['android:usesCleartextTraffic'] = 'true';
    return value;
  });
  return withInfoPlist(config, (value) => {
    value.modResults.NSLocalNetworkUsageDescription = '连接局域网中的 Dinosaur 媒体服务器';
    value.modResults.NSAppTransportSecurity = { NSAllowsArbitraryLoads: true };
    return value;
  });
};
