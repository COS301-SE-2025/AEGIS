import React from "react";
import { ChevronLeft, ChevronRight } from "lucide-react";
import { cn } from "../../lib/utils";

interface PaginationProps {
  page: number;
  totalPages: number;
  onChange: (page: number) => void;
}

export const Pagination: React.FC<PaginationProps> = ({
  page,
  totalPages,
  onChange,
}) => {
  if (totalPages <= 1) return null;

  const handlePrev = () => {
    if (page > 1) onChange(page - 1);
  };

  const handleNext = () => {
    if (page < totalPages) onChange(page + 1);
  };

  return (
    <div className="flex items-center gap-4">
      <button
        onClick={handlePrev}
        disabled={page === 1}
        className={cn(
          "px-3 py-1 rounded-md text-sm border",
          page === 1
            ? "text-muted-foreground border-muted cursor-not-allowed"
            : "hover:bg-muted"
        )}
      >
        <ChevronLeft size={16} />
      </button>
      <span className="text-sm text-muted-foreground">
        Page {page} of {totalPages}
      </span>
      <button
        onClick={handleNext}
        disabled={page === totalPages}
        className={cn(
          "px-3 py-1 rounded-md text-sm border",
          page === totalPages
            ? "text-muted-foreground border-muted cursor-not-allowed"
            : "hover:bg-muted"
        )}
      >
        <ChevronRight size={16} />
      </button>
    </div>
  );
};
