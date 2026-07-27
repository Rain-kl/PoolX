import { ArrowUp, SquarePen } from 'lucide-react';
import { useMemo, useState, type FormEvent, type KeyboardEvent } from 'react';
import { useTranslation } from 'react-i18next';

import { Button } from '@/components/ui/button';
import { Message, MessageContent, MessageGroup } from '@/components/ui/message';
import { Textarea } from '@/components/ui/textarea';
import {
  MOCK_CHAT_SEED,
  type ExampleChatMessage,
} from '@/features/example/mock-data';
import { PageHeader } from '@/shared/components/page-header';
import { cn } from '@/shared/lib/cn';

export function ExampleChatPage() {
  const { t } = useTranslation();
  const [messages, setMessages] =
    useState<ExampleChatMessage[]>(MOCK_CHAT_SEED);
  const [draft, setDraft] = useState('');
  const [sending, setSending] = useState(false);

  const canSend = draft.trim().length > 0 && !sending;

  const sortedMessages = useMemo(
    () => [...messages].sort((a, b) => a.createdAt.localeCompare(b.createdAt)),
    [messages],
  );

  function resetConversation(): void {
    setMessages(MOCK_CHAT_SEED);
    setDraft('');
  }

  function replyFor(prompt: string): string {
    const compact = prompt.trim().replace(/\s+/g, ' ');
    if (compact.length < 40) {
      return t('example.chat.mockReplyShort', { prompt: compact });
    }
    return t('example.chat.mockReplyLong', { prompt: compact.slice(0, 80) });
  }

  async function sendMessage(event?: FormEvent): Promise<void> {
    event?.preventDefault();
    const content = draft.trim();
    if (!content || sending) return;

    const userMessage: ExampleChatMessage = {
      id: `u-${Date.now()}`,
      role: 'user',
      content,
      createdAt: new Date().toISOString(),
    };
    setMessages((current) => [...current, userMessage]);
    setDraft('');
    setSending(true);

    await new Promise((resolve) => window.setTimeout(resolve, 450));
    const assistantMessage: ExampleChatMessage = {
      id: `a-${Date.now()}`,
      role: 'assistant',
      content: replyFor(content),
      createdAt: new Date().toISOString(),
    };
    setMessages((current) => [...current, assistantMessage]);
    setSending(false);
  }

  function onComposerKeyDown(event: KeyboardEvent<HTMLTextAreaElement>): void {
    if (event.key === 'Enter' && !event.shiftKey) {
      event.preventDefault();
      void sendMessage();
    }
  }

  return (
    <div className='flex min-h-[calc(100vh-10rem)] flex-col gap-5'>
      <PageHeader
        title={t('example.chat.title')}
        description={t('example.chat.description')}
        actions={
          <Button variant='secondary' size='sm' onClick={resetConversation}>
            <SquarePen />
            {t('example.chat.newConversation')}
          </Button>
        }
      />

      <div className='flex min-h-0 flex-1 flex-col overflow-hidden rounded-xl bg-card'>
        <div className='min-h-0 flex-1 space-y-4 overflow-y-auto px-4 py-5 sm:px-6'>
          {sortedMessages.length === 0 ? (
            <div className='flex h-full min-h-64 flex-col items-center justify-center text-center'>
              <p className='text-sm font-medium'>{t('example.chat.welcome')}</p>
              <p className='mt-2 max-w-sm text-xs text-muted-foreground'>
                {t('example.chat.welcomeHint')}
              </p>
            </div>
          ) : (
            <MessageGroup>
              {sortedMessages.map((message) => (
                <Message
                  key={message.id}
                  align={message.role === 'user' ? 'end' : 'start'}
                >
                  <MessageContent>
                    <div
                      className={cn(
                        'rounded-2xl px-3.5 py-2.5 text-sm leading-6',
                        message.role === 'user'
                          ? 'bg-primary text-primary-foreground'
                          : 'bg-secondary/70 text-foreground',
                      )}
                    >
                      {message.content}
                    </div>
                  </MessageContent>
                </Message>
              ))}
              {sending ? (
                <Message align='start'>
                  <MessageContent>
                    <div className='rounded-2xl bg-secondary/70 px-3.5 py-2.5 text-sm text-muted-foreground'>
                      {t('example.chat.thinking')}
                    </div>
                  </MessageContent>
                </Message>
              ) : null}
            </MessageGroup>
          )}
        </div>

        <form
          className='border-t border-border/60 p-4 sm:p-5'
          onSubmit={(event) => void sendMessage(event)}
        >
          <div className='overflow-hidden rounded-2xl bg-secondary/45 ring-1 ring-transparent transition-colors focus-within:bg-secondary/60 focus-within:ring-ring'>
            <Textarea
              value={draft}
              onChange={(event) => setDraft(event.target.value)}
              onKeyDown={onComposerKeyDown}
              placeholder={t('example.chat.placeholder')}
              className='min-h-24 resize-none border-0 bg-transparent px-4 py-3 text-sm shadow-none focus-visible:ring-0'
              aria-label={t('example.chat.placeholder')}
            />
            <div className='flex items-center justify-between gap-3 px-3 pb-3'>
              <p className='text-[11px] text-muted-foreground'>
                {t('example.chat.composerHint')}
              </p>
              <Button
                type='submit'
                size='icon'
                className='size-8'
                disabled={!canSend}
                aria-label={t('example.chat.send')}
              >
                <ArrowUp />
              </Button>
            </div>
          </div>
        </form>
      </div>
    </div>
  );
}
