"use client";

import { IconButton } from "@/components/helper";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { useAuth } from "@/context/auth";
import { TOAST_SUCCESS_STYLES } from "@/lib/utils";
import { useConnection, useWallet } from "@solana/wallet-adapter-react";
import { useWalletModal } from "@solana/wallet-adapter-react-ui";
import {
  LAMPORTS_PER_SOL,
  PublicKey,
  SystemProgram,
  Transaction,
} from "@solana/web3.js";
import { LogOut, Plus } from "lucide-react";
import React, { useRef, useState } from "react";
import { toast } from "sonner";
import { isDesktop } from "react-device-detect";
import { Checkbox } from "@/components/ui/checkbox";

export default function ClientHomePage() {
  const rechargeAmountRef = useRef<HTMLInputElement>(null);
  const identifierRef = useRef<HTMLInputElement>(null);
  const passwordRef = useRef<HTMLInputElement>(null);

  const [openRechargeModal, setOpenRechargeModal] = useState(false);
  const { user, underMaintenance, loginHandler } = useAuth();
  const [loading, setLoading] = useState(false);
  const [showLoginForm, setShowLoginForm] = useState(false);
  const [showPassword, setShowPassword] = useState(false);
  const [error, setError] = useState("");

  const wallet = useWallet();
  const walletModal = useWalletModal();
  const { connection } = useConnection();

  const addBalanceHandler = async () => {
    try {
      const amount = parseFloat(rechargeAmountRef.current?.value || "");
      if (!user || isNaN(amount) || !wallet.publicKey || !wallet.connected) {
        setError(
          isNaN(amount) ? "Please enter valid amount" : "Please connect wallet"
        );
        return;
      }

      setLoading(true);
      const to = process.env.NEXT_PUBLIC_KEY as string;
      const transaction = new Transaction();
      transaction.add(
        SystemProgram.transfer({
          fromPubkey: wallet.publicKey,
          toPubkey: new PublicKey(to),
          lamports: amount * LAMPORTS_PER_SOL,
        })
      );

      await wallet.sendTransaction(transaction, connection);
      toggleRechargeModal();
      setLoading(false);
      toast("Solana added successfully", {
        description: "Balance will be updated shortly",
        action: (
          <IconButton onClick={window.location.reload}>Refresh</IconButton>
        ),
        ...TOAST_SUCCESS_STYLES,
      });
    } catch (error) {
      if (error instanceof Error) {
        setLoading(false);
        setError(error.message);
      }
    }
  };

  const toggleLoginForm = () => setShowLoginForm((prev) => !prev);

  const toggleRechargeModal = () => {
    if (!loading) {
      setOpenRechargeModal((prev) => !prev);
    }
  };

  const login = () => {
    loginHandler(
      identifierRef.current?.value || "",
      passwordRef.current?.value || ""
    );
  };

  return (
    <>
      {wallet.connected ? (
        <div className="flex justify-between items-center  gap-2">
          <div className="text-primary">
            {(user?.solanaBalance || 0) / 10 ** 9} SOL
          </div>
          <Dialog open={openRechargeModal} onOpenChange={toggleRechargeModal}>
            <IconButton onClick={toggleRechargeModal}>
              <Plus size={18} />
              <span className="hidden md:inline-block">Add Sol</span>
            </IconButton>
            <DialogContent className=" bg-primary text-primary-foreground border-0 w-[90%] md:w-[60%] lg:w-[40%]">
              <DialogHeader>
                <DialogTitle className="">Add Solana</DialogTitle>

                <DialogDescription className="">
                  Note: In some cases it takes sometime for a transaction to get
                  distributed all over blockchain so verification might fail.
                  You have to retry the verification in those case.
                </DialogDescription>
              </DialogHeader>

              <div>
                <input
                  ref={rechargeAmountRef}
                  placeholder="Amount (2 SOL)"
                  type="number"
                  className="w-full  rounded-lg  bg-[#7c7c7c1f] px-4 py-2  active:outline-none focus:outline-none"
                />
                {error && (
                  <div className=" text-sm font-bold px-2 text-red-500 rounded-lg py-1">
                    {error}
                  </div>
                )}
                {underMaintenance && (
                  <div className=" text-sm font-bold px-2 text-red-500 rounded-lg py-1">
                    We are currently under maintenance so, please try to add
                    funds after sometime
                  </div>
                )}
              </div>
              <DialogFooter>
                <IconButton color="error" onClick={toggleRechargeModal}>
                  Cancel
                </IconButton>
                <IconButton
                  loading={loading}
                  disabled={underMaintenance}
                  onClick={addBalanceHandler}
                >
                  Submit
                </IconButton>
              </DialogFooter>
            </DialogContent>
          </Dialog>
          <IconButton
            color="error"
            onClick={() => {
              wallet.disconnect();
              localStorage.clear();
            }}
          >
            <LogOut size={20} />
            <span className="hidden md:inline-block">Disconnect</span>
          </IconButton>
        </div>
      ) : (
        <>
          {isDesktop ? (
            <button
              className="bg-[#2b9245] text-[#fff] text-xs md:text-sm font-bold  rounded-md px-5 py-2 mb-2"
              onClick={() => walletModal.setVisible(true)}
            >
              Connect
            </button>
          ) : (
            <Dialog open={showLoginForm} onOpenChange={toggleLoginForm}>
              <button
                className="bg-[#2b9245] text-[#fff] text-xs md:text-sm font-bold  rounded-md px-5 py-2 mb-2"
                onClick={toggleLoginForm}
              >
                Login
              </button>
              <DialogContent className=" bg-primary text-primary-foreground border-0 w-[90%] md:w-[60%] lg:w-[40%]">
                <DialogHeader>
                  <DialogTitle className="">Login</DialogTitle>

                  <DialogDescription className="">
                    Note: You can&apos;t register for this device you can only
                    register form desktop which has web3 wallet like Phantom,
                    Backpack etc.
                  </DialogDescription>
                </DialogHeader>

                <div>
                  <input
                    ref={identifierRef}
                    placeholder="Public key"
                    className="w-full mb-2 rounded-lg  bg-[#7c7c7c1f] px-4 py-2  active:outline-none focus:outline-none"
                  />
                  <input
                    ref={passwordRef}
                    type={showPassword ? "text" : "password"}
                    placeholder="Password"
                    className="w-full  rounded-lg  bg-[#7c7c7c1f] px-4 py-2  active:outline-none focus:outline-none"
                  />
                  <div className="cursor-pointer flex items-center space-x-2 mt-2">
                    <Checkbox
                      checked={showPassword}
                      onCheckedChange={(v) => setShowPassword(v as boolean)}
                      id="terms"
                      className="dark"
                    />
                    <label
                      htmlFor="terms"
                      className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
                    >
                      Show Password
                    </label>
                  </div>
                  {error && (
                    <div className=" text-sm font-bold px-2 text-red-500 rounded-lg py-1">
                      {error}
                    </div>
                  )}
                </div>
                <DialogFooter>
                  <IconButton color="error" onClick={toggleLoginForm}>
                    Cancel
                  </IconButton>
                  <IconButton
                    loading={loading}
                    // disabled={underMaintenance}
                    onClick={login}
                  >
                    Login
                  </IconButton>
                </DialogFooter>
              </DialogContent>
            </Dialog>
          )}
        </>
      )}
    </>
  );
}
