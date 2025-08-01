class Logger {
    constructor(enableConsole = false) {
        this.enableConsole = enableConsole;
        this.logs = [];
        this.maxLogs = 1000;
    }

    _log(level, message, data = null) {
        const timestamp = new Date().toISOString();
        const logEntry = { timestamp, level, message, data };

        this.logs.push(logEntry);
        if (this.logs.length > this.maxLogs) {
            this.logs.shift();
        }

        if (this.enableConsole) {
            const args = [message];
            if (data) args.push(data);

            switch (level) {
                case 'error':
                    console.error(...args);
                    break;
                case 'warn':
                    console.warn(...args);
                    break;
                case 'info':
                    console.info(...args);
                    break;
                case 'debug':
                default:
                    console.log(...args);
                    break;
            }
        }
    }

    error(message, data) {
        this._log('error', message, data);
    }

    warn(message, data) {
        this._log('warn', message, data);
    }

    info(message, data) {
        this._log('info', message, data);
    }

    debug(message, data) {
        this._log('debug', message, data);
    }

    getLogs(count = 50) {
        return this.logs.slice(-count);
    }

    clearLogs() {
        this.logs = [];
    }

    setConsoleOutput(enabled) {
        this.enableConsole = enabled;
    }
}

const isDevelopment = window.location.hostname === 'localhost' || window.location.hostname.includes('127.0.0.1');

export const logger = new Logger(isDevelopment);

window.logger = logger;
