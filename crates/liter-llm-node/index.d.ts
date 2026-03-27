/* auto-generated type stubs for @kreuzberg/liter-llm-native */

export interface LlmClientOptions {
	apiKey: string;
	baseUrl?: string;
	modelHint?: string;
	maxRetries?: number;
	timeoutSecs?: number;
}

export class LlmClient {
	constructor(options: LlmClientOptions);
	chat(request: Record<string, unknown>): Promise<Record<string, unknown>>;
	chatStream(request: Record<string, unknown>): Promise<Record<string, unknown>[]>;
	embed(request: Record<string, unknown>): Promise<Record<string, unknown>>;
	listModels(): Promise<Record<string, unknown>>;
	imageGenerate(request: Record<string, unknown>): Promise<Record<string, unknown>>;
	speech(request: Record<string, unknown>): Promise<Buffer>;
	transcribe(request: Record<string, unknown>): Promise<Record<string, unknown>>;
	moderate(request: Record<string, unknown>): Promise<Record<string, unknown>>;
	rerank(request: Record<string, unknown>): Promise<Record<string, unknown>>;
	createFile(request: Record<string, unknown>): Promise<Record<string, unknown>>;
	retrieveFile(fileId: string): Promise<Record<string, unknown>>;
	deleteFile(fileId: string): Promise<Record<string, unknown>>;
	listFiles(query?: Record<string, unknown>): Promise<Record<string, unknown>>;
	fileContent(fileId: string): Promise<Buffer>;
	createBatch(request: Record<string, unknown>): Promise<Record<string, unknown>>;
	retrieveBatch(batchId: string): Promise<Record<string, unknown>>;
	listBatches(query?: Record<string, unknown>): Promise<Record<string, unknown>>;
	cancelBatch(batchId: string): Promise<Record<string, unknown>>;
	createResponse(request: Record<string, unknown>): Promise<Record<string, unknown>>;
	retrieveResponse(id: string): Promise<Record<string, unknown>>;
	cancelResponse(id: string): Promise<Record<string, unknown>>;
}

export function version(): string;
