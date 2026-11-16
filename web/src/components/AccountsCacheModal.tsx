import React from 'react';
import { X, Database, ShieldCheck, Building, ExternalLink } from 'lucide-react';
import { AccountRecord } from '../types';

interface AccountsCacheModalProps {
  isOpen: boolean;
  onClose: () => void;
  accounts: AccountRecord[];
}

export const AccountsCacheModal: React.FC<AccountsCacheModalProps> = ({
  isOpen,
  onClose,
  accounts,
}) => {
  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-950/80 backdrop-blur-sm animate-in fade-in duration-150">
      <div className="bg-slate-900 border border-slate-800 rounded-xl w-full max-w-3xl max-h-[85vh] flex flex-col shadow-2xl overflow-hidden">
        {/* Header */}
        <div className="p-4 border-b border-slate-800 flex items-center justify-between bg-slate-900/90">
          <div className="flex items-center gap-2">
            <div className="w-7 h-7 rounded-lg bg-purple-500/10 border border-purple-500/30 flex items-center justify-center text-purple-400">
              <Database className="w-4 h-4" />
            </div>
            <div>
              <h3 className="font-bold text-sm text-white">CRM Account Deduplication Index</h3>
              <p className="text-xs text-slate-400">In-memory SQLite cached accounts preventing duplicate outreach</p>
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
        <div className="flex-1 overflow-y-auto p-4">
          {accounts.length === 0 ? (
            <div className="py-12 text-center text-slate-500 text-xs">
              No accounts currently indexed in cache. Ingest leads to seed the deduplication index.
            </div>
          ) : (
            <div className="border border-slate-800 rounded-lg overflow-hidden">
              <table className="w-full text-left text-xs">
                <thead className="bg-slate-950/70 border-b border-slate-800 text-slate-400 font-semibold">
                  <tr>
                    <th className="p-3">Company Name</th>
                    <th className="p-3">Corporate Domain</th>
                    <th className="p-3">Assigned AE Owner</th>
                    <th className="p-3">CRM ID</th>
                    <th className="p-3">Source</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-800/60 bg-slate-900/40">
                  {accounts.map((acc) => (
                    <tr key={acc.id} className="hover:bg-slate-800/40 transition">
                      <td className="p-3 font-medium text-white flex items-center gap-1.5">
                        <Building className="w-3.5 h-3.5 text-purple-400" />
                        {acc.company_name}
                      </td>
                      <td className="p-3 font-mono text-emerald-400">{acc.domain}</td>
                      <td className="p-3 text-slate-300">
                        {acc.owner_name}
                        {acc.owner_email && (
                          <span className="text-[10px] text-slate-500 block">{acc.owner_email}</span>
                        )}
                      </td>
                      <td className="p-3 font-mono text-slate-400">{acc.crm_id}</td>
                      <td className="p-3">
                        <span className="text-[10px] px-2 py-0.5 rounded bg-purple-500/10 text-purple-300 border border-purple-500/20 font-mono uppercase">
                          {acc.crm_type}
                        </span>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};
