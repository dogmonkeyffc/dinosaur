const { withAppBuildGradle } = require('expo/config-plugins');

module.exports = function withSharedInputs(config) {
  return withAppBuildGradle(config, (value) => {
    const marker = '// Dinosaur shared TypeScript bundle inputs';
    if (!value.modResults.contents.includes(marker)) {
      value.modResults.contents += `
${marker}
tasks.matching { it.name.startsWith("createBundle") && it.name.endsWith("JsAndAssets") }.configureEach {
    inputs.files(fileTree(rootProject.file("../../../packages")) {
        include "**/src/**", "**/package.json"
        exclude "**/node_modules/**"
    })
}
`;
    }
    return value;
  });
};
