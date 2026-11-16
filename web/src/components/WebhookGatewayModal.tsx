import React, { useState, useEffect } from 'react';
import { X, Globe, Copy, Check, Terminal, Code } from 'lucide-react';
import { WebhookInfo } from '../types';

interface WebhookGatewayModalProps {
  isOpen: boolean;
  onClose: () => void;
}

export const WebhookGatewayModal: React.FC<WebhookGatewayModalProps> = ({ isOpen, onClose }) => {
  const [info, setInfo] = useState<WebhookInfo | null>(null);
  const [copiedIndex, setCopiedIndex] = useState<number | null>(null);

  useEffect(() => {
    if (isOpen) {
      fetch('/api/v1/webhooks/info')
        .then((res) => res.json())
        .then((data) => setInfo(data))
        .catch(() => {});
    }
  }, [isOpen]);

  if (!isOpen) return null;

  const handleCopyCurl = (cmd: string, index: number) => {
    navigator.clipboard.writeText(cmd);
    setCopiedIndex(index);
    setTimeout(() => setCopiedIndex(null), 2000);
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-950/80 backdrop-blur-sm animate-in fade-in duration-150">
      <div className="bg-slate-900 border border-slate-800 rounded-xl w-full max-w-3xl max-h-[90vh] flex flex-col shadow-2xl overflow-hidden">
        {/* Header */}
        <div className="p-4 border-b border-slate-800 flex items-center justify-between bg-slate-900/90">
          <div className="flex items-center gap-2">
            <div className="w-7 h-7 rounded-lg bg-emerald-500/10 border border-emerald-500/30 flex items-center justify-center text-emerald-400">
              <Globe className="w-4 h-4" />
            </div>
            <div>
              <h3 className="font-bold text-sm text-white">Inbound Webhook Gateway & cURL Playground</h3>
              <p className="text-xs text-slate-400">Production endpoints for SendGrid, Postmark, SES, Mailgun, and custom forms</p>
            </div>
          </div>
          <button
            onClick={onClose}
            className="p-1 text-slate-400 hover:text-slate-200 bg-slate-800 hover:bg-slate-700 rounded-md transition"
          >
            <X className="w-4 h-4" />
          </button>
        </div>

        {/* Content */}
        <div className="flex-1 overflow-y-auto p-5 space-y-4">
          {info?.webhooks.map((wh, idx) => (
            <div key={idx} className="bg-slate-950 border border-slate-800 rounded-lg p-3.5 space-y-2.5">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <span className="font-bold text-xs text-white">{wh.provider}</span>
                  <span className="text-[10px] font-mono px-2 py-0.5 rounded bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                    {wh.method}
                  </span>
                  <span className="text-[10px] text-slate-500 font-mono">{wh.content_type}</span>
                </div>
                <button
                  onClick={() => handleCopyCurl(wh.curl_sample, idx)}
                  className="flex items-center gap-1.5 px-2.5 py-1 text-[11px] font-medium bg-slate-800 hover:bg-slate-700 text-slate-300 rounded border border-slate-700 transition"
                >
                  {copiedIndex === idx ? (
                    <>
                      <Check className="w-3 h-3 text-emerald-400" />
                      <span className="text-emerald-400">Copied!</span>
                    </>
                  ) : (
                    <>
                      <Copy className="w-3 h-3 text-slate-400" />
                      <span>Copy cURL</span>
                    </>
                  )}
                </button>
              </div>

              <div className="text-xs font-mono text-emerald-400/90 bg-slate-900/80 px-2.5 py-1.5 rounded border border-slate-800">
                {wh.endpoint}
              </div>

              <div className="relative group">
                <pre className="p-2.5 bg-slate-900 text-slate-300 font-mono text-[10px] rounded overflow-x-auto whitespace-pre-wrap leading-relaxed border border-slate-800/80">
                  {wh.curl_sample}
                </pre>
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
};
