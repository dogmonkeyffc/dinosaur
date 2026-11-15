const { AndroidConfig, withInfoPlist } = require('expo/config-plugins');

module.exports = function withDiscovery(config) {
  config = AndroidConfig.Permissions.withPermissions(config, [
    'android.permission.INTERNET',
    'android.permission.ACCESS_NETWORK_STATE',
    'android.permission.ACCESS_WIFI_STATE',
    'android.permission.CHANGE_WIFI_MULTICAST_STATE',
  ]);
  return withInfoPlist(config, (value) => {
    const services = value.modResults.NSBonjourServices || [];
    value.modResults.NSBonjourServices = [...new Set([...services, '_dinosaur._tcp'])];
    return value;
  });
};
