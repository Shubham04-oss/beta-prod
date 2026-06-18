import { X, ShoppingBag, CreditCard, Hash, Clock, MoreHorizontal, CheckCircle2, type LucideIcon } from "lucide-react";
import { Button } from "./ui/button";

export type OrderSidebarRecord = {
  id: string;
  status: string;
  currency: string;
  subtotal: number;
  discount_total: number;
  shipping_total: number;
  tax_total: number;
  total: number;
  payment_status?: string | null;
  payment_provider?: string | null;
  payment_reference?: string | null;
  channel?: string | null;
  source_platform?: string | null;
  idempotency_key?: string | null;
  created_at: string;
  updated_at: string;
};

type OrderDetailsSidebarProps = {
  order?: OrderSidebarRecord | null;
  onClose?: () => void;
};

export function OrderDetailsSidebar({ order, onClose }: OrderDetailsSidebarProps) {
  const createdAt = order?.created_at ? new Date(order.created_at) : null;

  return (
    <div className="flex flex-col h-full w-full">
      <div className="flex items-center justify-between mb-8">
        <h2 className="font-semibold text-lg leading-tight tracking-tight">Order Details</h2>
        <button onClick={onClose} className="text-zinc-500 hover:text-white transition-colors" aria-label="Close order details">
          <X className="w-5 h-5" />
        </button>
      </div>

      {!order ? (
        <div className="flex-1 flex flex-col items-center justify-center text-center px-6">
          <div className="w-12 h-12 rounded-full bg-zinc-900 border border-zinc-800 flex items-center justify-center mb-4">
            <ShoppingBag className="w-5 h-5 text-zinc-500" />
          </div>
          <p className="text-sm font-medium text-zinc-200">Select an order</p>
          <p className="text-xs text-zinc-500 mt-2">Order context, payment state, and event timing appear here.</p>
        </div>
      ) : (
        <>
          <div className="flex-1 overflow-y-auto pr-2 custom-scrollbar flex flex-col gap-8 pb-10">
            <div className="flex items-center gap-4">
              <div className="w-12 h-12 rounded-full bg-blue-500/10 border border-blue-500/20 flex items-center justify-center flex-shrink-0">
                <ShoppingBag className="w-5 h-5 text-blue-400" />
              </div>
              <div className="min-w-0">
                <div className="flex items-center gap-2">
                  <h3 className="font-bold text-base text-zinc-100 truncate">{shortOrderID(order.id)}</h3>
                  <span className="bg-blue-500/20 text-blue-400 text-[10px] font-bold px-2 py-0.5 rounded-full capitalize">
                    {order.status.replaceAll("_", " ")}
                  </span>
                </div>
                <p className="text-[11px] text-zinc-400 mt-1">
                  {createdAt ? createdAt.toLocaleString() : "No timestamp"}
                </p>
              </div>
            </div>

            <section>
              <h4 className="text-xs font-semibold text-zinc-500 uppercase tracking-wider mb-3">Order Summary</h4>
              <div className="bg-zinc-900/50 rounded-2xl p-4 border border-zinc-800 flex flex-col gap-2">
                <SummaryRow label="Subtotal" value={money(order.subtotal, order.currency)} />
                <SummaryRow label="Discounts" value={money(order.discount_total, order.currency)} />
                <SummaryRow label="Shipping" value={money(order.shipping_total, order.currency)} />
                <SummaryRow label="Tax" value={money(order.tax_total, order.currency)} />
                <div className="h-px w-full bg-zinc-800 my-1" />
                <div className="flex justify-between text-sm font-bold">
                  <span className="text-zinc-200">Total</span>
                  <span className="text-white">{money(order.total, order.currency)}</span>
                </div>
              </div>
            </section>

            <section>
              <h4 className="text-xs font-semibold text-zinc-500 uppercase tracking-wider mb-3">Payment</h4>
              <div className="grid grid-cols-2 gap-2">
                <InfoTile icon={CreditCard} label="Status" value={order.payment_status || "Pending"} />
                <InfoTile icon={CreditCard} label="Provider" value={order.payment_provider || "Not set"} />
                <InfoTile icon={Hash} label="Reference" value={order.payment_reference || "Not set"} wide />
              </div>
            </section>

            <section>
              <h4 className="text-xs font-semibold text-zinc-500 uppercase tracking-wider mb-3">Channel</h4>
              <div className="grid grid-cols-2 gap-2">
                <InfoTile icon={Hash} label="Channel" value={order.channel || order.source_platform || "Direct"} />
                <InfoTile icon={Clock} label="Updated" value={order.updated_at ? new Date(order.updated_at).toLocaleString() : "No timestamp"} />
                <InfoTile icon={Hash} label="Idempotency" value={order.idempotency_key || "Not set"} wide />
              </div>
            </section>

            <section>
              <h4 className="text-xs font-semibold text-zinc-500 uppercase tracking-wider mb-3">Order Timeline</h4>
              <div className="relative pl-6 pb-2 border-l border-zinc-800 ml-3">
                <div className="absolute w-5 h-5 bg-green-500 rounded-full -left-2.5 top-0 border-4 border-[#121212] flex items-center justify-center">
                  <CheckCircle2 className="w-3 h-3 text-white" />
                </div>
                <div className="-mt-1">
                  <p className="text-xs font-medium text-zinc-200">Order persisted</p>
                  <p className="text-[10px] text-zinc-500 mt-0.5">{createdAt ? createdAt.toLocaleString() : "No timestamp"}</p>
                </div>
              </div>
            </section>
          </div>

          <div className="pt-4 flex gap-2 border-t border-zinc-800/50 mt-auto">
            <Button variant="outline" className="flex-1 bg-transparent border-red-900/30 text-red-400 hover:bg-red-950/30 hover:text-red-300 h-10 rounded-xl">
              Cancel Order
            </Button>
            <Button variant="outline" className="flex-1 bg-zinc-900 border-zinc-800 text-zinc-200 hover:bg-zinc-800 hover:text-white h-10 rounded-xl">
              Edit Order
            </Button>
            <Button variant="outline" size="icon" className="w-10 h-10 bg-zinc-900 border-zinc-800 text-zinc-400 hover:bg-zinc-800 hover:text-white rounded-xl flex-shrink-0">
              <MoreHorizontal className="w-4 h-4" />
            </Button>
          </div>
        </>
      )}
    </div>
  );
}

function SummaryRow({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex justify-between text-xs">
      <span className="text-zinc-400">{label}</span>
      <span className="text-zinc-200 font-medium">{value}</span>
    </div>
  );
}

function InfoTile({ icon: Icon, label, value, wide }: { icon: LucideIcon; label: string; value: string; wide?: boolean }) {
  return (
    <div className={`bg-zinc-900/50 rounded-xl p-3 border border-zinc-800 min-w-0 ${wide ? "col-span-2" : ""}`}>
      <Icon className="w-3.5 h-3.5 text-zinc-500 mb-2" />
      <p className="text-[10px] text-zinc-500">{label}</p>
      <p className="text-xs font-medium text-zinc-200 mt-0.5 truncate">{value}</p>
    </div>
  );
}

function shortOrderID(id: string) {
  return id ? `#${id.slice(0, 8).toUpperCase()}` : "#ORDER";
}

function money(value: number, currency: string) {
  return new Intl.NumberFormat("en-US", { style: "currency", currency: currency || "USD" }).format(value || 0);
}
