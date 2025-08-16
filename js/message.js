/**
 * 工作线程消息类型定义
 * 用于在主线程和工作线程之间进行通信
 */

/**
 * 消息类型枚举
 * 预分配常用消息类型对象，避免重复创建
 */
const MESSAGE_TYPES = Object.freeze({
  // 初始化相关消息
  INIT_REQUEST: Object.freeze({ type: "initialize" }),
  INIT_ERROR: Object.freeze({ type: "initialization-error" }),
  INIT_SUCCESS: Object.freeze({ type: "initialization-success" }),
  
  // 工作线程状态消息
  WORKER_READY: Object.freeze({ type: "worker-ready" }),
  
  // 翻译相关消息
  TRANSLATION_REQUEST: Object.freeze({ type: "translation-request" }),
  TRANSLATION_RESPONSE: Object.freeze({ type: "translation-response" }),
  TRANSLATION_ERROR: Object.freeze({ type: "translation-error" }),
  
  // 队列管理消息
  TRANSLATIONS_DISCARDED: Object.freeze({ type: "translations-discarded" }),
  DISCARD_TRANSLATION_QUEUE: Object.freeze({ type: "discard-translation-queue" }),
  CANCEL_SINGLE_TRANSLATION: Object.freeze({ type: "cancel-single-translation" }),
});

/**
 * 消息类型字符串常量
 * 用于类型检查和调试
 */
const MESSAGE_TYPE_STRINGS = Object.freeze({
  INITIALIZE: "initialize",
  INITIALIZATION_ERROR: "initialization-error",
  INITIALIZATION_SUCCESS: "initialization-success",
  WORKER_READY: "worker-ready",
  TRANSLATION_REQUEST: "translation-request",
  TRANSLATION_RESPONSE: "translation-response",
  TRANSLATION_ERROR: "translation-error",
  TRANSLATIONS_DISCARDED: "translations-discarded",
  DISCARD_TRANSLATION_QUEUE: "discard-translation-queue",
  CANCEL_SINGLE_TRANSLATION: "cancel-single-translation",
});

/**
 * 验证消息类型是否有效
 * @param {string} messageType - 要验证的消息类型
 * @returns {boolean} 是否为有效的消息类型
 */
function isValidMessageType(messageType) {
  return Object.values(MESSAGE_TYPE_STRINGS).includes(messageType);
}

/**
 * 创建标准消息对象
 * @param {string} type - 消息类型
 * @param {any} [payload] - 消息载荷
 * @param {string} [id] - 消息ID
 * @returns {Object} 标准消息对象
 */
function createMessage(type, payload = null, id = null) {
  const message = { type };
  
  if (payload !== null) {
    message.payload = payload;
  }
  
  if (id !== null) {
    message.id = id;
  }
  
  return message;
}

module.exports = {
  MESSAGE_TYPES,
  MESSAGE_TYPE_STRINGS,
  isValidMessageType,
  createMessage,
};
