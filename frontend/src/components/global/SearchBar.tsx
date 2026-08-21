import React, { useState } from "react";
import { Input } from "antd";
import { SearchOutlined } from "@ant-design/icons";

export interface SearchBarProps {
  value?: string;
  placeholder?: string;
  onSearch?: (value: string) => void;
  onChange?: (value: string) => void;
  className?: string;
  style?: React.CSSProperties;
  allowClear?: boolean;
  disabled?: boolean;
  size?: "small" | "middle" | "large";
}

export const SearchBar: React.FC<SearchBarProps> = ({
  value: externalValue,
  placeholder = "ค้นหา...",
  onSearch,
  onChange,
  className,
  style,
  allowClear = true,
  disabled = false,
  size,
}) => {
  const [internalValue, setInternalValue] = useState(externalValue ?? "");


  const value = externalValue ?? internalValue;

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const newValue = e.target.value;
    setInternalValue(newValue);
    onChange?.(newValue);
  };

  const handleSearch = () => {
    onSearch?.(value.trim());
  };

  return (
    <Input
      allowClear={allowClear}
      disabled={disabled}
      className={className}
      style={style}
      size={size}
      prefix={<SearchOutlined className="text-black/45" />}
      placeholder={placeholder}
      value={value}
      onChange={handleChange}
      onPressEnter={handleSearch}
    />
  );
};

export default SearchBar;
