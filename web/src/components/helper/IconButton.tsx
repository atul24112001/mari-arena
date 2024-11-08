import React, { PropsWithChildren } from "react";
import Spinner from "./Spinner";

export function IconButton({
  children,
  onClick,
  color = "success",
  disabled,
  loading,
}: PropsWithChildren<Props>) {
  const colors = {
    success: "text-[#2b9245] hover:bg-[#2b924520] ",
    error: "text-[#d43939] hover:bg-[#d4393920]",
  };
  return (
    <button
      disabled={disabled}
      className={`${colors[color]} flex gap-1 text-xs md:text-sm font-bold  p-2 rounded-full `}
      onClick={() => {
        if (!loading) {
          onClick?.();
        }
      }}
    >
      {loading ? <Spinner /> : children}
    </button>
  );
}

type Props = {
  onClick?: () => void;
  color?: "success" | "error";
  disabled?: boolean;
  loading?: boolean;
};
