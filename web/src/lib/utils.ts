import { clsx, type ClassValue } from "clsx";
import { twMerge } from "tailwind-merge";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

export const TOAST_ERROR_STYLES = {
  className: "bg-primary text-red-500  border-0",
  descriptionClassName: "text-primary-foreground",
};

export const TOAST_SUCCESS_STYLES = {
  className: "bg-primary text-white  border-0",
  // descriptionClassName: "text-primary-foreground",
};
