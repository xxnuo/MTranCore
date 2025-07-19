// 预先分配常用消息类型
const MESSAGE_TYPES = {
  INIT_REQUEST: { type: "initialize" },
  INIT_ERROR: { type: "initialization-error" },
  INIT_SUCCESS: { type: "initialization-success" },
  TRANSLATIONS_DISCARDED: { type: "translations-discarded" },
  WORKER_READY: { type: "worker-ready" },
  TRANSLATION_REQUEST: { type: "translation-request" },
  TRANSLATION_RESPONSE: { type: "translation-response" },
  TRANSLATION_ERROR: { type: "translation-error" },
  DISCARD_TRANSLATION_QUEUE: { type: "discard-translation-queue" },
  CANCEL_SINGLE_TRANSLATION: { type: "cancel-single-translation" },
};

module.exports = {
  MESSAGE_TYPES,
};
