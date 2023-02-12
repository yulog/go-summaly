export class StatusError extends Error {
    constructor(message, statusCode, statusMessage) {
        super(message);
        this.name = 'StatusError';
        this.statusCode = statusCode;
        this.statusMessage = statusMessage;
        this.isPermanentError = typeof this.statusCode === 'number' && this.statusCode >= 400 && this.statusCode < 500;
    }
}
