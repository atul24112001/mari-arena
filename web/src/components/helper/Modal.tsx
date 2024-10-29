import { X } from "lucide-react";
import React, { PropsWithChildren } from "react";
import { IconButton } from "./IconButton";

export function Modal({
  title,
  onClose,
  open,
  children,
  className,
  error,
}: PropsWithChildren<Props>) {
  return (
    <div
      className={`${
        open ? "" : "hidden"
      } fixed top-0 text-[#2b9245] bottom-0 left-0 right-0 bg-[#00000020] flex justify-center items-center`}
    >
      <div className="bg-[#e9f5ed] w-[90%] md:w-[60%] lg:w-[40%] px-5 py-3 rounded-lg">
        <div className="flex justify-between items-center">
          <h2 className="font-bold text-xl">{title || ""}</h2>
          <IconButton color="error" onClick={onClose}>
            <X size={20} />
          </IconButton>
        </div>
        <div className={className}>
          {error && (
            <div className="mb-3 text-sm font-bold bg-red-100 text-red-500 rounded-lg px-4 py-1">
              {error}
            </div>
          )}
          {children}
        </div>
      </div>
    </div>
  );
}

type Props = {
  title: string;
  open: boolean;
  onClose: () => void;
  className?: string;
  error: string;
};
