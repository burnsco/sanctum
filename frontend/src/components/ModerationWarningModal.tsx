import { AlertTriangle, Ban, ShieldAlert } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { useModerationWarningStore } from "@/stores/useModerationWarningStore";

const MAX_STRIKES = 3;

export function ModerationWarningModal() {
  const { open, data, dismiss } = useModerationWarningStore();

  if (!data) return null;

  const { strikes, isBanned, message } = data;
  const remaining = Math.max(MAX_STRIKES - strikes, 0);

  return (
    <Dialog open={open} onOpenChange={(open) => !open && dismiss()}>
      <DialogContent
        className="sm:max-w-md border-destructive/50 bg-background"
        onInteractOutside={(e) => {
          // Force user to acknowledge by clicking the button
          if (isBanned) e.preventDefault();
        }}
      >
        <DialogHeader className="gap-2">
          <div className="mx-auto flex h-16 w-16 items-center justify-center rounded-full bg-destructive/15 ring-4 ring-destructive/20">
            {isBanned ? (
              <Ban className="h-8 w-8 text-destructive" />
            ) : (
              <ShieldAlert className="h-8 w-8 text-destructive" />
            )}
          </div>

          <DialogTitle className="text-center text-xl font-bold text-destructive">
            {isBanned ? "Account Banned" : "Content Policy Violation"}
          </DialogTitle>
        </DialogHeader>

        <div className="space-y-4 text-center">
          {!isBanned && (
            <div className="mx-auto flex w-fit items-center gap-1.5 rounded-full border border-destructive/40 bg-destructive/10 px-4 py-1.5">
              {Array.from({ length: MAX_STRIKES }, (_, strikeNumber) => strikeNumber + 1).map(
                (strikeNumber) => (
                  <span
                    key={strikeNumber}
                    className={`h-3 w-3 rounded-full transition-colors ${
                      strikeNumber <= strikes ? "bg-destructive" : "bg-muted-foreground/30"
                    }`}
                  />
                ),
              )}
              <span className="ml-2 text-sm font-semibold text-destructive">
                Strike {strikes} of {MAX_STRIKES}
              </span>
            </div>
          )}

          <p className="text-sm text-muted-foreground">
            {isBanned ? (
              <>
                Your account has been{" "}
                <span className="font-semibold text-destructive">permanently banned</span> after{" "}
                <span className="font-semibold">{MAX_STRIKES} violations</span> of our community
                guidelines.
              </>
            ) : (
              <>
                Your content was removed because it violates our community guidelines.{" "}
                {remaining > 0 ? (
                  <>
                    You have{" "}
                    <span className="font-semibold text-destructive">
                      {remaining} strike{remaining !== 1 ? "s" : ""} remaining
                    </span>{" "}
                    before your account is banned.
                  </>
                ) : (
                  <>This is your final warning.</>
                )}
              </>
            )}
          </p>

          {message && (
            <div className="rounded-lg border border-border bg-muted/50 px-4 py-2.5 text-left text-xs text-muted-foreground">
              <span className="flex items-center gap-1.5 font-medium text-foreground">
                <AlertTriangle className="h-3.5 w-3.5" />
                Reason
              </span>
              <p className="mt-1 capitalize">{message}</p>
            </div>
          )}

          {isBanned && (
            <p className="text-xs text-muted-foreground">
              If you believe this was a mistake, contact support.
            </p>
          )}
        </div>

        <Button variant={isBanned ? "destructive" : "default"} className="w-full" onClick={dismiss}>
          {isBanned ? "I understand" : "Acknowledge warning"}
        </Button>
      </DialogContent>
    </Dialog>
  );
}
