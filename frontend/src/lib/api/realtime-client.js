/**
 * リアルタイム更新クライアント
 * WebSocketとポーリングの両方をサポートし、自動的にフォールバックする
 */

import { writable } from 'svelte/store';

// 接続状態の定数
const CONNECTION_STATES = {
    DISCONNECTED: 'disconnected',
    CONNECTING: 'connecting',
    CONNECTED: 'connected',
    RECONNECTING: 'reconnecting',
    FAILED: 'failed'
};

// メッセージタイプの定数
const MESSAGE_TYPES = {
    CONNECT: 'connect',
    DISCONNECT: 'disconnect',
    AUTH: 'auth',
    SUBSCRIBE: 'subscribe',
    UNSUBSCRIBE: 'unsubscribe',
    TOURNAMENT_UPDATE: 'tournament_update',
    MATCH_UPDATE: 'match_update',
    MATCH_RESULT: 'match_result',
    BRACKET_UPDATE: 'bracket_update',
    ERROR: 'error',
    PING: 'ping',
    PONG: 'pong'
};

class RealtimeClient {
    constructor(options = {}) {
        this.options = {
            wsUrl: options.wsUrl || 'ws://localhost:8080/ws',
            apiUrl: options.apiUrl || 'http://localhost:8080/api/v1',
            pollingInterval: options.pollingInterval || 30000, // 30秒
            maxReconnectAttempts: options.maxReconnectAttempts || 5,
            reconnectDelay: options.reconnectDelay || 1000,
            useWebSocket: options.useWebSocket !== false, // デフォルトでWebSocket使用
            usePolling: options.usePolling !== false, // デフォルトでポーリング使用
            debug: options.debug || false,
            ...options
        };

        // 状態管理
        this.connectionState = writable(CONNECTION_STATES.DISCONNECTED);
        this.isAuthenticated = writable(false);
        this.subscribedSports = writable([]);
        this.lastError = writable(null);

        // 内部状態
        this.ws = null;
        this.token = null;
        this.reconnectAttempts = 0;
        this.reconnectTimer = null;
        this.pollingTimers = new Map();
        this.pollingETags = new Map();
        this.eventListeners = new Map();
        this.currentSports = [];
        this.isWebSocketSupported = true;
        this.isPollingActive = false;

        // WebSocket使用可能性をチェック
        this.checkWebSocketSupport();
    }

    /**
     * WebSocketサポートをチェック
     */
    checkWebSocketSupport() {
        this.isWebSocketSupported = typeof WebSocket !== 'undefined' && this.options.useWebSocket;
        if (!this.isWebSocketSupported) {
            this.log('WebSocket is not supported or disabled, using polling only');
        }
    }

    /**
     * 接続を開始
     */
    async connect(token = null) {
        this.token = token;
        
        if (this.isWebSocketSupported) {
            await this.connectWebSocket();
        } else {
            await this.startPolling();
        }
    }

    /**
     * WebSocket接続を開始
     */
    async connectWebSocket() {
        if (this.ws && this.ws.readyState === WebSocket.OPEN) {
            return;
        }

        this.connectionState.set(CONNECTION_STATES.CONNECTING);
        this.log('Connecting to WebSocket...');

        try {
            this.ws = new WebSocket(this.options.wsUrl);
            this.setupWebSocketEventHandlers();
        } catch (error) {
            this.log('WebSocket connection failed:', error);
            this.handleWebSocketError(error);
        }
    }

    /**
     * WebSocketイベントハンドラーを設定
     */
    setupWebSocketEventHandlers() {
        this.ws.onopen = () => {
            this.log('WebSocket connected');
            this.connectionState.set(CONNECTION_STATES.CONNECTED);
            this.reconnectAttempts = 0;
            
            // 認証を実行
            if (this.token) {
                this.authenticate(this.token);
            }
        };

        this.ws.onmessage = (event) => {
            try {
                const message = JSON.parse(event.data);
                this.handleWebSocketMessage(message);
            } catch (error) {
                this.log('Failed to parse WebSocket message:', error);
            }
        };

        this.ws.onclose = (event) => {
            this.log('WebSocket closed:', event.code, event.reason);
            this.connectionState.set(CONNECTION_STATES.DISCONNECTED);
            
            if (event.code !== 1000) { // 正常終了以外
                this.handleWebSocketError(new Error(`WebSocket closed with code ${event.code}`));
            }
        };

        this.ws.onerror = (error) => {
            this.log('WebSocket error:', error);
            this.handleWebSocketError(error);
        };
    }

    /**
     * WebSocketメッセージを処理
     */
    handleWebSocketMessage(message) {
        this.log('Received WebSocket message:', message);

        switch (message.type) {
            case MESSAGE_TYPES.CONNECT:
                this.log('WebSocket connection confirmed');
                break;

            case MESSAGE_TYPES.AUTH:
                if (message.data && message.data.success) {
                    this.isAuthenticated.set(true);
                    this.log('WebSocket authentication successful');
                    
                    // 既存の購読を復元
                    if (this.currentSports.length > 0) {
                        this.subscribe(this.currentSports);
                    }
                } else {
                    this.isAuthenticated.set(false);
                    this.log('WebSocket authentication failed');
                }
                break;

            case MESSAGE_TYPES.SUBSCRIBE:
                if (message.data && message.data.success) {
                    this.subscribedSports.set(message.data.sports || []);
                    this.log('Subscription successful:', message.data.sports);
                }
                break;

            case MESSAGE_TYPES.TOURNAMENT_UPDATE:
            case MESSAGE_TYPES.MATCH_UPDATE:
            case MESSAGE_TYPES.MATCH_RESULT:
            case MESSAGE_TYPES.BRACKET_UPDATE:
                this.emitEvent(message.type, message.data);
                break;

            case MESSAGE_TYPES.ERROR:
                this.handleError(message.data);
                break;

            case MESSAGE_TYPES.PING:
                this.sendPong();
                break;

            default:
                this.log('Unknown message type:', message.type);
        }
    }

    /**
     * WebSocketエラーを処理
     */
    handleWebSocketError(error) {
        this.lastError.set(error.message);
        
        if (this.reconnectAttempts < this.options.maxReconnectAttempts) {
            this.scheduleReconnect();
        } else {
            this.log('Max reconnect attempts reached, falling back to polling');
            this.fallbackToPolling();
        }
    }

    /**
     * 再接続をスケジュール
     */
    scheduleReconnect() {
        this.connectionState.set(CONNECTION_STATES.RECONNECTING);
        this.reconnectAttempts++;
        
        const delay = this.options.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1);
        this.log(`Scheduling reconnect attempt ${this.reconnectAttempts} in ${delay}ms`);
        
        this.reconnectTimer = setTimeout(() => {
            this.connectWebSocket();
        }, delay);
    }

    /**
     * ポーリングにフォールバック
     */
    async fallbackToPolling() {
        this.log('Falling back to polling mode');
        this.isWebSocketSupported = false;
        await this.startPolling();
    }

    /**
     * ポーリングを開始
     */
    async startPolling() {
        if (!this.options.usePolling) {
            this.log('Polling is disabled');
            this.connectionState.set(CONNECTION_STATES.FAILED);
            return;
        }

        this.log('Starting polling mode');
        this.isPollingActive = true;
        this.connectionState.set(CONNECTION_STATES.CONNECTED);
        this.isAuthenticated.set(true); // ポーリングでは認証不要

        // 購読中のスポーツに対してポーリングを開始
        for (const sport of this.currentSports) {
            this.startPollingForSport(sport);
        }
    }

    /**
     * 特定のスポーツに対してポーリングを開始
     */
    startPollingForSport(sport) {
        const dataTypes = ['tournament', 'matches', 'bracket'];
        
        for (const dataType of dataTypes) {
            const key = `${sport}:${dataType}`;
            
            if (this.pollingTimers.has(key)) {
                continue; // 既にポーリング中
            }

            const poll = async () => {
                try {
                    const lastETag = this.pollingETags.get(key);
                    const response = await this.checkForUpdates(sport, dataType, lastETag);
                    
                    if (response.has_updates) {
                        this.pollingETags.set(key, response.etag);
                        this.handlePollingUpdate(sport, dataType, response.data);
                    }
                    
                    // 次のポーリングをスケジュール
                    const interval = response.next_poll_seconds * 1000 || this.options.pollingInterval;
                    this.pollingTimers.set(key, setTimeout(poll, interval));
                    
                } catch (error) {
                    this.log(`Polling error for ${key}:`, error);
                    // エラー時は通常の間隔でリトライ
                    this.pollingTimers.set(key, setTimeout(poll, this.options.pollingInterval));
                }
            };

            // 初回実行
            poll();
        }
    }

    /**
     * ポーリング更新を処理
     */
    handlePollingUpdate(sport, dataType, data) {
        let messageType;
        
        switch (dataType) {
            case 'tournament':
                messageType = MESSAGE_TYPES.TOURNAMENT_UPDATE;
                break;
            case 'matches':
                messageType = MESSAGE_TYPES.MATCH_UPDATE;
                break;
            case 'bracket':
                messageType = MESSAGE_TYPES.BRACKET_UPDATE;
                break;
            default:
                return;
        }

        this.emitEvent(messageType, {
            sport: sport,
            data: data,
            action: 'updated',
            timestamp: new Date().toISOString()
        });
    }

    /**
     * 更新をチェック（ポーリング用）
     */
    async checkForUpdates(sport, dataType, lastETag = null) {
        const url = new URL(`${this.options.apiUrl}/polling/${sport}/${dataType}/check`);
        if (lastETag) {
            url.searchParams.set('last_etag', lastETag);
        }

        const response = await fetch(url.toString());
        if (!response.ok) {
            throw new Error(`Polling request failed: ${response.status}`);
        }

        const result = await response.json();
        return result.data;
    }

    /**
     * 認証を実行
     */
    authenticate(token) {
        if (this.isWebSocketSupported && this.ws && this.ws.readyState === WebSocket.OPEN) {
            this.sendWebSocketMessage(MESSAGE_TYPES.AUTH, { token });
        }
        // ポーリングモードでは認証不要
    }

    /**
     * スポーツを購読
     */
    subscribe(sports) {
        if (!Array.isArray(sports)) {
            sports = [sports];
        }

        this.currentSports = [...new Set([...this.currentSports, ...sports])];

        if (this.isWebSocketSupported && this.ws && this.ws.readyState === WebSocket.OPEN) {
            this.sendWebSocketMessage(MESSAGE_TYPES.SUBSCRIBE, { sports });
        } else if (this.isPollingActive) {
            // ポーリングモードでは新しいスポーツのポーリングを開始
            for (const sport of sports) {
                this.startPollingForSport(sport);
            }
            this.subscribedSports.set(this.currentSports);
        }
    }

    /**
     * スポーツの購読を解除
     */
    unsubscribe(sports) {
        if (!Array.isArray(sports)) {
            sports = [sports];
        }

        this.currentSports = this.currentSports.filter(sport => !sports.includes(sport));

        if (this.isWebSocketSupported && this.ws && this.ws.readyState === WebSocket.OPEN) {
            this.sendWebSocketMessage(MESSAGE_TYPES.UNSUBSCRIBE, { sports });
        } else if (this.isPollingActive) {
            // ポーリングタイマーを停止
            for (const sport of sports) {
                const dataTypes = ['tournament', 'matches', 'bracket'];
                for (const dataType of dataTypes) {
                    const key = `${sport}:${dataType}`;
                    const timer = this.pollingTimers.get(key);
                    if (timer) {
                        clearTimeout(timer);
                        this.pollingTimers.delete(key);
                        this.pollingETags.delete(key);
                    }
                }
            }
            this.subscribedSports.set(this.currentSports);
        }
    }

    /**
     * WebSocketメッセージを送信
     */
    sendWebSocketMessage(type, data = {}) {
        if (this.ws && this.ws.readyState === WebSocket.OPEN) {
            const message = {
                type,
                data,
                timestamp: new Date().toISOString()
            };
            this.ws.send(JSON.stringify(message));
            this.log('Sent WebSocket message:', message);
        }
    }

    /**
     * Pongメッセージを送信
     */
    sendPong() {
        this.sendWebSocketMessage(MESSAGE_TYPES.PONG, {
            timestamp: new Date().toISOString()
        });
    }

    /**
     * イベントリスナーを追加
     */
    addEventListener(type, listener) {
        if (!this.eventListeners.has(type)) {
            this.eventListeners.set(type, []);
        }
        this.eventListeners.get(type).push(listener);
    }

    /**
     * イベントリスナーを削除
     */
    removeEventListener(type, listener) {
        const listeners = this.eventListeners.get(type);
        if (listeners) {
            const index = listeners.indexOf(listener);
            if (index > -1) {
                listeners.splice(index, 1);
            }
        }
    }

    /**
     * イベントを発火
     */
    emitEvent(type, data) {
        const listeners = this.eventListeners.get(type);
        if (listeners) {
            listeners.forEach(listener => {
                try {
                    listener(data);
                } catch (error) {
                    this.log('Event listener error:', error);
                }
            });
        }
    }

    /**
     * エラーを処理
     */
    handleError(error) {
        this.log('Received error:', error);
        this.lastError.set(error.message || error.toString());
        this.emitEvent('error', error);
    }

    /**
     * 接続を切断
     */
    disconnect() {
        this.log('Disconnecting...');

        // WebSocket接続を閉じる
        if (this.ws) {
            this.ws.close(1000, 'Client disconnect');
            this.ws = null;
        }

        // 再接続タイマーをクリア
        if (this.reconnectTimer) {
            clearTimeout(this.reconnectTimer);
            this.reconnectTimer = null;
        }

        // ポーリングタイマーをクリア
        for (const timer of this.pollingTimers.values()) {
            clearTimeout(timer);
        }
        this.pollingTimers.clear();
        this.pollingETags.clear();

        // 状態をリセット
        this.isPollingActive = false;
        this.reconnectAttempts = 0;
        this.currentSports = [];
        this.connectionState.set(CONNECTION_STATES.DISCONNECTED);
        this.isAuthenticated.set(false);
        this.subscribedSports.set([]);
    }

    /**
     * 接続状態を取得
     */
    getConnectionState() {
        return this.connectionState;
    }

    /**
     * 認証状態を取得
     */
    getAuthenticationState() {
        return this.isAuthenticated;
    }

    /**
     * 購読中のスポーツを取得
     */
    getSubscribedSports() {
        return this.subscribedSports;
    }

    /**
     * 最後のエラーを取得
     */
    getLastError() {
        return this.lastError;
    }

    /**
     * デバッグログを出力
     */
    log(...args) {
        if (this.options.debug) {
            console.log('[RealtimeClient]', ...args);
        }
    }
}

// シングルトンインスタンス
let realtimeClient = null;

/**
 * リアルタイムクライアントのインスタンスを取得
 */
export function getRealtimeClient(options = {}) {
    if (!realtimeClient) {
        realtimeClient = new RealtimeClient(options);
    }
    return realtimeClient;
}

/**
 * リアルタイムクライアントを初期化
 */
export function initRealtimeClient(options = {}) {
    realtimeClient = new RealtimeClient(options);
    return realtimeClient;
}

export { RealtimeClient, CONNECTION_STATES, MESSAGE_TYPES };