"use client";

import React, { useEffect, useState } from "react";
import { SettingsLeftSidebar } from "@/components/SettingsLeftSidebar";
import { SettingsInsightsSidebar } from "@/components/SettingsInsightsSidebar";
import { UserPlus, MoreHorizontal, ShieldCheck, Mail, Bot, KeyRound } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { fetchAPI } from "@/lib/api";
import { useAuth } from "@/providers/AuthProvider";
import { auth } from "@/lib/firebase";
import { sendPasswordResetEmail } from "firebase/auth";

interface Member {
  id: string;
  email: string;
  role: string;
  created_at: string;
}

export default function OrganizationSettingsPage() {
  const { role } = useAuth();
  const [members, setMembers] = useState<Member[]>([]);
  const [loading, setLoading] = useState(true);
  const [inviteEmail, setInviteEmail] = useState("");
  const [inviteRole, setInviteRole] = useState("MANAGER");
  const [customRoleName, setCustomRoleName] = useState("");
  const [inviting, setInviting] = useState(false);
  const [inviteMessage, setInviteMessage] = useState<{ text: string; type: "success" | "error" } | null>(null);
  const [isInviteModalOpen, setIsInviteModalOpen] = useState(false);

  useEffect(() => {
    fetchMembers();
  }, []);

  const fetchMembers = async () => {
    try {
      setLoading(true);
      const data = await fetchAPI("/api/v1/organization/members");
      if (data) {
        setMembers(data as Member[]);
      }
    } catch (err) {
      console.error("Failed to load members:", err);
    } finally {
      setLoading(false);
    }
  };

  const handleInvite = async (e: React.FormEvent) => {
    e.preventDefault();
    setInviting(true);
    setInviteMessage(null);
    try {
      const finalRole = inviteRole === "CUSTOM" ? customRoleName.toUpperCase() : inviteRole;
      await fetchAPI("/api/v1/organization/members", {
        method: "POST",
        body: JSON.stringify({ email: inviteEmail, role: finalRole }),
      });
      setInviteMessage({ text: "Member invited successfully! A password reset link has been sent.", type: "success" });
      setInviteEmail("");
      setIsInviteModalOpen(false);
      fetchMembers(); // Refresh the list
    } catch (err: any) {
      console.error("Failed to invite member:", err);
      setInviteMessage({ text: err.message || "Failed to invite member. Check permissions.", type: "error" });
    } finally {
      setInviting(false);
    }
  };

  const handleResetPassword = async (email: string) => {
    try {
      await sendPasswordResetEmail(auth, email);
      setInviteMessage({ text: `Password reset email sent to ${email}`, type: "success" });
    } catch (err: any) {
      console.error("Failed to send reset email:", err);
      setInviteMessage({ text: err.message || "Failed to send reset email.", type: "error" });
    }
  };

  return (
    <>
      {/* Left Sidebar */}
      <SettingsLeftSidebar />

      {/* Main Content Area */}
      <main className="flex-1 overflow-y-auto custom-scrollbar pr-4 flex flex-col gap-8 pb-10">
        
        {/* Header */}
        <div className="mt-2 flex justify-between items-end">
          <div>
            <h1 className="text-3xl font-bold tracking-tight">Organization</h1>
            <p className="text-muted-foreground mt-1">Manage your team members, roles, and permissions.</p>
          </div>
          {role === "ADMIN" && (
            <Button 
              onClick={() => setIsInviteModalOpen(!isInviteModalOpen)}
              className="h-10 bg-black text-white hover:bg-black/90 dark:bg-white dark:text-black dark:hover:bg-white/90 rounded-full"
            >
              <UserPlus className="w-4 h-4 mr-2" />
              Invite Member
            </Button>
          )}
        </div>

        {inviteMessage && !isInviteModalOpen && (
          <div className={`p-3 rounded-md text-sm ${inviteMessage.type === "success" ? "bg-emerald-50 text-emerald-600 dark:bg-emerald-950/50 dark:text-emerald-400" : "bg-red-50 text-red-600 dark:bg-red-950/50 dark:text-red-400"}`}>
            {inviteMessage.text}
          </div>
        )}

        {/* Invite Member Section (Inline Modal Equivalent) */}
        {isInviteModalOpen && role === "ADMIN" && (
          <div className="bg-white/20 dark:bg-white/5 rounded-[2rem] p-6 border border-white/30 dark:border-purple-500/30 flex flex-col gap-4 animate-in fade-in slide-in-from-top-4">
            <div>
              <h3 className="font-semibold text-[15px]">Invite New Team Member</h3>
              <p className="text-[11px] text-muted-foreground mt-0.5">They will receive an email to set their password and join your workspace.</p>
            </div>

            {inviteMessage && (
               <div className={`p-3 rounded-md text-sm ${inviteMessage.type === "success" ? "bg-emerald-50 text-emerald-600" : "bg-red-50 text-red-600"}`}>
                 {inviteMessage.text}
               </div>
            )}

            <form onSubmit={handleInvite} className="flex items-end gap-4">
              <div className="flex flex-col gap-1.5 flex-1">
                <label className="text-[11px] font-medium text-muted-foreground">Email Address</label>
                <div className="relative">
                  <Mail className="absolute left-3 top-3 h-4 w-4 text-slate-400" />
                  <input 
                    type="email" 
                    placeholder="colleague@yourcompany.com" 
                    className="w-full h-11 pl-9 rounded-xl bg-white/40 dark:bg-black/20 border border-black/5 dark:border-white/10 text-[13px] font-medium focus:outline-none focus:ring-2 focus:ring-purple-500/50" 
                    value={inviteEmail}
                    onChange={(e) => setInviteEmail(e.target.value)}
                    required
                  />
                </div>
              </div>
              <div className="flex flex-col gap-1.5 w-48">
                <label className="text-[11px] font-medium text-muted-foreground">Role</label>
                <select 
                  className="w-full h-11 px-3 rounded-xl bg-white/40 dark:bg-black/20 border border-black/5 dark:border-white/10 text-[13px] focus:outline-none"
                  value={inviteRole}
                  onChange={(e) => setInviteRole(e.target.value)}
                >
                  <option value="MANAGER">Manager</option>
                  <option value="ADMIN">Admin</option>
                  <option value="CUSTOM">Custom Role...</option>
                </select>
              </div>
              
              {inviteRole === "CUSTOM" && (
                <div className="flex flex-col gap-1.5 w-48 animate-in fade-in slide-in-from-left-2">
                  <label className="text-[11px] font-medium text-muted-foreground">Custom Role Name</label>
                  <input 
                    type="text" 
                    placeholder="e.g. EDITOR" 
                    className="w-full h-11 px-3 rounded-xl bg-white/40 dark:bg-black/20 border border-black/5 dark:border-white/10 text-[13px] font-medium focus:outline-none focus:ring-2 focus:ring-purple-500/50" 
                    value={customRoleName}
                    onChange={(e) => setCustomRoleName(e.target.value)}
                    required
                  />
                </div>
              )}

              <Button type="submit" className="h-11 bg-black text-white hover:bg-black/90 dark:bg-white dark:text-black dark:hover:bg-white/90 rounded-full px-6" disabled={inviting || (inviteRole === "CUSTOM" && !customRoleName)}>
                {inviting ? "Sending..." : "Send Invite"}
              </Button>
              <Button type="button" variant="ghost" className="h-11" onClick={() => setIsInviteModalOpen(false)}>
                Cancel
              </Button>
            </form>
          </div>
        )}

        {/* Members Table */}
        <div className="bg-white/20 dark:bg-white/5 rounded-[2rem] border border-white/30 dark:border-white/10 overflow-hidden">
          <div className="p-6 border-b border-black/5 dark:border-white/5 flex justify-between items-center">
            <div>
              <h3 className="font-semibold text-[15px]">Team Members</h3>
              <p className="text-[11px] text-muted-foreground mt-0.5">All active users assigned to your workspace.</p>
            </div>
            <div className="text-xs font-medium bg-purple-100 text-purple-700 dark:bg-purple-900/30 dark:text-purple-400 px-3 py-1 rounded-full">
              {members.length} Active
            </div>
          </div>
          
          <div className="overflow-x-auto">
            <table className="w-full text-left text-[13px]">
              <thead className="bg-black/5 dark:bg-white/5 text-muted-foreground font-medium">
                <tr>
                  <th className="px-6 py-3">Member</th>
                  <th className="px-6 py-3">Role</th>
                  <th className="px-6 py-3">Joined</th>
                  <th className="px-6 py-3 text-right">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-black/5 dark:divide-white/5">
                {loading ? (
                  <tr>
                    <td colSpan={4} className="px-6 py-12 text-center text-muted-foreground">
                      <div className="flex flex-col items-center justify-center">
                        <div className="w-6 h-6 border-2 border-purple-600 border-t-transparent rounded-full animate-spin mb-2"></div>
                        Loading members...
                      </div>
                    </td>
                  </tr>
                ) : members.length === 0 ? (
                  <tr>
                    <td colSpan={4} className="px-6 py-12 text-center text-muted-foreground">
                      No members found in this workspace.
                    </td>
                  </tr>
                ) : (
                  members.map((member) => (
                    <tr key={member.id} className="hover:bg-white/40 dark:hover:bg-white/5 transition-colors group">
                      <td className="px-6 py-4">
                        <div className="flex items-center gap-3">
                          <div className="w-8 h-8 rounded-full bg-gradient-to-tr from-purple-500 to-blue-500 flex items-center justify-center text-white font-bold text-xs uppercase">
                            {member.email.charAt(0)}
                          </div>
                          <div>
                            <div className="font-medium text-foreground">{member.email.split('@')[0]}</div>
                            <div className="text-[11px] text-muted-foreground">{member.email}</div>
                          </div>
                        </div>
                      </td>
                      <td className="px-6 py-4">
                        <div className="flex items-center gap-1.5">
                          {member.role === 'ADMIN' && <ShieldCheck className="w-3.5 h-3.5 text-amber-500" />}
                          {member.role !== 'ADMIN' && <Bot className="w-3.5 h-3.5 text-blue-500" />}
                          <span className={`px-2 py-0.5 rounded-md text-[10px] font-bold uppercase ${
                            member.role === 'ADMIN' ? 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400' :
                            member.role === 'MANAGER' ? 'bg-purple-100 text-purple-700 dark:bg-purple-900/30 dark:text-purple-400' :
                            'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400'
                          }`}>
                            {member.role}
                          </span>
                        </div>
                      </td>
                      <td className="px-6 py-4 text-muted-foreground">
                        {new Date(member.created_at).toLocaleDateString()}
                      </td>
                      <td className="px-6 py-4 text-right">
                        <Popover>
                          <PopoverTrigger className="inline-flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground hover:text-foreground opacity-0 group-hover:opacity-100 transition-opacity hover:bg-black/5 dark:hover:bg-white/5 focus:outline-none">
                            <MoreHorizontal className="w-4 h-4" />
                          </PopoverTrigger>
                          <PopoverContent align="end" className="w-48 p-2 rounded-2xl bg-white/80 dark:bg-black/60 backdrop-blur-xl border border-black/5 dark:border-white/10 shadow-xl">
                            <button 
                              onClick={() => handleResetPassword(member.email)}
                              className="w-full flex items-center gap-2 px-2 py-2 text-sm text-foreground hover:bg-black/5 dark:hover:bg-white/5 rounded-xl transition-colors"
                            >
                              <KeyRound className="w-4 h-4 text-purple-500" />
                              Reset Password
                            </button>
                          </PopoverContent>
                        </Popover>
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        </div>
      </main>

      {/* Right Matte Black Area */}
      <aside className="w-[360px] flex-shrink-0 bg-[#121212] text-white p-6 rounded-3xl relative z-20 flex flex-col h-full shadow-[0_8px_30px_rgb(0,0,0,0.4)] border border-white/5">
        <SettingsInsightsSidebar />
      </aside>
    </>
  );
}
