import { h } from 'vue';
import i18n from '@/i18n';

let antFeedbackModulePromise: Promise<typeof import('ant-design-vue')> | null = null;

function loadAntFeedbackModule() {
  antFeedbackModulePromise ??= import('ant-design-vue');
  return antFeedbackModulePromise;
}

async function copyTextToClipboard(value: string): Promise<void> {
  const text = String(value ?? '');
  let clipboardError: unknown;

  if (typeof navigator !== 'undefined' && navigator.clipboard?.writeText) {
    try {
      await navigator.clipboard.writeText(text);
      return;
    } catch (error) {
      clipboardError = error;
    }
  }

  if (copyTextWithExecCommand(text)) {
    return;
  }

  throw clipboardError ?? new Error('Clipboard is unavailable');
}

function copyTextWithExecCommand(text: string): boolean {
  if (typeof document === 'undefined' || !document.body) {
    return false;
  }

  const textarea = document.createElement('textarea');
  const selection = document.getSelection();
  const activeElement = document.activeElement instanceof HTMLElement ? document.activeElement : null;
  const range = selection && selection.rangeCount > 0 ? selection.getRangeAt(0) : null;

  textarea.value = text;
  textarea.setAttribute('readonly', '');
  textarea.style.position = 'fixed';
  textarea.style.top = '0';
  textarea.style.left = '0';
  textarea.style.width = '1px';
  textarea.style.height = '1px';
  textarea.style.padding = '0';
  textarea.style.border = '0';
  textarea.style.opacity = '0';
  textarea.style.pointerEvents = 'none';

  document.body.appendChild(textarea);
  textarea.focus();
  textarea.select();
  textarea.setSelectionRange(0, textarea.value.length);

  let copied = false;
  try {
    copied = document.execCommand('copy');
  } finally {
    document.body.removeChild(textarea);

    if (selection) {
      selection.removeAllRanges();
      if (range) {
        selection.addRange(range);
      }
    }

    activeElement?.focus();
  }

  return copied;
}

export function showSuccess(content: string) {
  void loadAntFeedbackModule().then(({ message }) => {
    message.success(content);
  });
}

export function showError(content: string) {
  void loadAntFeedbackModule().then(({ message }) => {
    message.error(content);
  });
}

export function showWarning(content: string) {
  void loadAntFeedbackModule().then(({ message }) => {
    message.warning(content);
  });
}

export function showInfo(content: string) {
  void loadAntFeedbackModule().then(({ message }) => {
    message.info(content);
  });
}

export function showConfirm(config: Record<string, any>) {
  void loadAntFeedbackModule().then(({ Modal }) => {
    Modal.confirm(config);
  });
}

export function showWarningModal(config: Record<string, any>) {
  void loadAntFeedbackModule().then(({ Modal }) => {
    Modal.warning(config);
  });
}

export function showCopyValueModal(config: {
  title: string;
  value: string;
  message?: string;
  width?: number;
}) {
  void loadAntFeedbackModule().then(({ Modal, Button, Input, Space }) => {
    const value = config.value || '';
    const copyValue = async () => {
      try {
        await copyTextToClipboard(value);
        showSuccess(i18n.global.t('misc.copySuccess'));
      } catch {
        showError(i18n.global.t('misc.copyError'));
      }
    };

    Modal.info({
      title: config.title,
      width: config.width ?? 560,
      okText: i18n.global.t('misc.close'),
      content: h(
        Space,
        {
          direction: 'vertical',
          size: 12,
          style: {
            width: '100%',
          },
        },
        {
          default: () => [
            config.message
              ? h(
                'div',
                {
                  class: 'portal-copy-modal__message',
                },
                config.message,
              )
              : null,
            h(Input.TextArea, {
              value,
              autoSize: {
                minRows: 2,
                maxRows: 4,
              },
              readonly: true,
            }),
            h(
              Button,
              {
                type: 'primary',
                onClick: copyValue,
              },
              {
                default: () => i18n.global.t('misc.copy'),
              },
            ),
          ],
        },
      ),
    });
  });
}
