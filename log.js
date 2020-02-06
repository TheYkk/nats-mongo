cobf = {name: "auth-ms",
  level: "debug",
  useLevelLabels: true,}
  const logger = require('pino')(cobf);

setInterval(() => {
  logger.error(`User not found , code 555 , ${Math.random()}`);
}, 2222);
