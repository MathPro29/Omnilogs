import { Descriptions, Empty, Tag } from "antd";
import type { SearchRecord } from "@/features/log-explorer/services/log-search.service";
import {
  compactRoutePattern,
  fieldGroups,
  getCustomFieldValue,
  getLogAuthor,
  getLogErrorInfo,
  getValue,
  maskSensitive,
} from "@/features/log-explorer/utils/log-utils";

const missing = "-";
const formatTimestamp = (value: string) => {
  if (!value || value === missing) return value;
  const parsed = new Date(value);
  if (Number.isNaN(parsed.getTime())) return value;
  return parsed.toLocaleString("sv-SE", { timeZone: "Asia/Bangkok", hour12: false });
};

const documentLabels: Record<string, string> = {
  water: "บิลค่าน้ำ",
  central: "บิลค่าส่วนกลาง",
  insurance: "บิลประกัน",
  other: "บิลอื่น ๆ",
  clean: "รายการทำความสะอาด",
  receipt: "ใบเสร็จ",
  invoice: "ใบแจ้งหนี้",
  deposit: "เงินมัดจำ",
  fine: "ค่าปรับ",
  wallet: "กระเป๋าเงิน",
  debit_credit_note: "ใบลดหนี้/เพิ่มหนี้",
};

const businessValue = (record: SearchRecord, ...keys: string[]) =>
  getValue(
    record,
    ...keys.flatMap((key) => [
      key,
      "custom_fields." + key,
      "payload.custom_fields." + key,
      "data.custom_fields." + key,
      "metadata." + key,
      "payload.metadata." + key,
      "data.metadata." + key,
    ]),
  );

const prettyFieldValue = (value: unknown): string => {
  const masked = maskSensitive(value);
  if (masked === null || masked === undefined) return "-";
  return typeof masked === "object" ? JSON.stringify(masked) : String(masked);
};

const flattenFields = (
  value: unknown,
  prefix = "",
  result: Array<[string, string]> = [],
) => {
  if (value === null || value === undefined) return result;
  if (Array.isArray(value)) {
    value.forEach((item, index) =>
      flattenFields(item, prefix + "[" + index + "]", result),
    );
    return result;
  }
  if (typeof value !== "object") {
    result.push([prefix || "value", prettyFieldValue(value)]);
    return result;
  }
  const entries = Object.entries(value as Record<string, unknown>);
  if (!entries.length) return result;
  entries.forEach(([key, child]) =>
    flattenFields(child, prefix ? prefix + "." + key : key, result),
  );
  return result;
};

export function SummaryCard({ record }: { record: SearchRecord }) {
  const errorInfo = getLogErrorInfo(record);
  const routePattern = getValue(
    record,
    "route_pattern",
    "payload.route_pattern",
    "data.route_pattern",
  );
  const author = getLogAuthor(record);
  const actionOperation = getValue(
    record,
    "metadata.action.operation",
    "payload.metadata.action.operation",
    "data.metadata.action.operation",
  );
  const actionMethod = getValue(
    record,
    "metadata.action.method",
    "payload.metadata.action.method",
    "data.metadata.action.method",
    "method",
    "payload.request_method",
  );
  const actionPath = getValue(
    record,
    "metadata.action.resource_path",
    "payload.metadata.action.resource_path",
    "data.metadata.action.resource_path",
    "path",
    "payload.request_path",
  );
  const action =
    actionOperation !== missing
      ? actionOperation
      : [actionMethod, actionPath]
          .filter((value) => value !== missing)
          .join(" ") || missing;
  const documentType = getCustomFieldValue(record, "type");
  const documentLabel = documentLabels[documentType] ?? documentType;
  const documentID = businessValue(
    record,
    "billWatersId",
    "billCentralsId",
    "billInsurancesId",
    "billOthersId",
    "invoicesId",
    "receiptsId",
    "cleansId",
  );
  const year = businessValue(record, "year");
  const month = businessValue(record, "month");
  const billingPeriod =
    year !== missing && month !== missing
      ? `${month}/${year}`
      : year !== missing
        ? year
        : month;











  const businessItems = [
    ["เอกสาร", documentLabel],
    ["เลขเอกสาร/บิล", documentID],
    ["สมาชิก/ผู้รับบริการ", businessValue(record, "membersId", "members_id")],
    ["ห้อง/ยูนิต", businessValue(record, "addressesId", "addresses_id")],
    ["อาคาร", businessValue(record, "buildingsId", "buildings_id")],
    ["งวดบิล", billingPeriod],
    [
      "วันที่เอกสาร",
      businessValue(record, "askTime", "confirmTime", "dueTime", "createdAt"),
    ],
    ["ยอดเงิน", businessValue(record, "realPrice", "price", "amount")],
    ["การดำเนินการ", action],
  ].filter(([, value]) => value !== missing);
  const items = [
    ["Action", action],
    ["Document type", getCustomFieldValue(record, "type")],
    [
      "Timestamp",
      formatTimestamp(getValue(record, "@timestamp", "timestamp", "payload.timestamp")),
    ],
    ["Level", getValue(record, "level", "payload.log_level", "payload.level")],
    [
      "Status",
      getValue(
        record,
        "payload.status_code",
        "status_code",
        "payload.status",
        "status",
      ),
    ],
    [
      "Method",
      getValue(
        record,
        "method",
        "payload.method",
        "payload.request_method",
        "data.request_method",
        "request.method",
      ),
    ],
    [
      "API path",
      getValue(
        record,
        "path",
        "request_path",
        "payload.request_path",
        "data.request_path",
        "request.path",
      ),
    ],
    ["Route pattern", compactRoutePattern(routePattern)],
    [
      "Product",
      getValue(
        record,
        "product_name",
        "payload.product",
        "payload.product_name",
        "product",
        "service",
        "product_id",
      ),
    ],
    [
      "Environment",
      getValue(
        record,
        "payload.environment",
        "payload.environment_name",
        "environment",
        "environment_id",
      ),
    ],
    [
      "Source project ID",
      getValue(
        record,
        "source_project_id",
        "payload.source_project_id",
        "data.source_project_id",
        "projectsId",
      ),
    ],
    [
      "Project",
      getValue(
        record,
        "project_name",
        "project_code",
        "project_id",
        "payload.project_id",
        "data.project_id",
      ),
    ],
    [
      "Classification",
      getValue(
        record,
        "classification_status",
        "payload.classification_status",
        "data.classification_status",
        "routing_status",
      ),
    ],
    [
      "Hierarchy",
      getValue(
        record,
        "feature_full_path",
        "payload.feature_full_path",
        "data.feature_full_path",
        "raw.feature_full_path",
        "raw.payload.feature_full_path",
      ),
    ],
    [
      "Routing",
      getValue(
        record,
        "routing_method",
        "payload.routing_method",
        "data.routing_method",
      ),
    ],
    [
      "Routing status",
      getValue(
        record,
        "routing_status",
        "payload.routing_status",
        "data.routing_status",
      ),
    ],
    [
      "Route key",
      getValue(record, "route_key", "payload.route_key", "data.route_key"),
    ],
    [
      "Routing reason",
      getValue(
        record,
        "routing_reason",
        "payload.routing_reason",
        "data.routing_reason",
      ),
    ],
    [
      "Trace ID",
      getValue(
        record,
        "payload.trace_id",
        "trace_id",
        "payload.correlation_id",
        "correlation_id",
      ),
    ],
    ["Author ID", author.id ?? missing],
    ["Author name", author.name ?? missing],
    ["Author role", author.role ?? missing],
    [
      "IP",
      getValue(record, "payload.ip", "payload.client_ip", "ip", "client_ip"),
    ],
    [
      "Project ID",
      getValue(
        record,
        "metadata.actor.project_id",
        "payload.metadata.actor.project_id",
        "data.metadata.actor.project_id",
        "actor.project_id",
        "payload.actor.project_id",
        "data.actor.project_id",
        "custom_fields.project_id",
        "payload.custom_fields.project_id",
        "metadata.project_id",
        "payload.metadata.project_id",
        "payload.project_id",
        "data.project_id",
        "project_id",
      ),
    ],
    [
      "Project Name",
      getValue(
        record,
        "project_name",
        "payload.project_name",
        "data.project_name",
      ),
    ],
  ];
  const error = getValue(
    record,
    "payload.error_message",
    "error_message",
    "payload.error",
    "error",
  );

  const renderFieldValue = (label: string, value: string) => {
    if (label === "Level" && value !== missing) {
      const upper = value.toUpperCase();
      if (upper === "WARN" || upper === "WARNING")
        return <Tag color="#D4B106">WARN</Tag>;
      if (
        upper === "ERROR" ||
        upper === "FATAL" ||
        upper === "CRITICAL" ||
        upper === "ERR"
      )
        return <Tag color="#FF3D89">ERROR</Tag>;
      if (upper === "INFO") return <Tag color="#1890FF">INFO</Tag>;
      return <Tag>{value}</Tag>;
    }
    if (label === "Status" && value !== missing) {
      const num = Number(value);
      if (!isNaN(num) && num > 0) {
        const text = errorInfo.statusText ?? "";
        const labelText = text ? `${num} (${text})` : String(num);
        if (num >= 500) return <Tag color="#FF3D89">HTTP {labelText}</Tag>;
        if (num >= 400) return <Tag color="#D4B106">HTTP {labelText}</Tag>;
        if (num >= 200 && num < 300)
          return <Tag color="#00BE8F">HTTP {labelText}</Tag>;
        return <Tag color="default">HTTP {labelText}</Tag>;
      }
    }
    return value;
  };

  return (
    <section className="log-summary" aria-label="Log summary">
      {businessItems.length > 0 && (
        <section
          className="log-summary__business"
          aria-label="Business summary"
        >
          <h3 className="p-5">สรุปรายการ</h3>
          <div className="log-summary__grid">
            {businessItems.map(([label, value]) => (
              <div key={label}>
                <dt>{label}</dt>
                <dd title={value}>{value}</dd>
              </div>
            ))}
          </div>

    </section>
      )}
      <div className="log-summary__grid">
        {items.map(([label, value]) => (
          <div key={label}>
            <dt>{label}</dt>
            <dd title={value}>{renderFieldValue(label, value)}</dd>
          </div>
        ))}
      </div>
      {error !== missing && (
        <div className="log-summary__error">
          <dt>Error message</dt>
          <dd>{error}</dd>
        </div>
      )}
    </section>
  );
}

export function PrettyLogView({ record }: { record: SearchRecord }) {
  const type = getCustomFieldValue(record, "type");
  const action = getValue(
    record,
    "metadata.action.operation",
    "payload.metadata.action.operation",
    "data.metadata.action.operation",
    "path",
    "payload.request_path",
  );
  const year = businessValue(record, "year");
  const month = businessValue(record, "month");
  const period =
    year !== missing && month !== missing ? month + "/" + year : missing;
  const featureCode = getCustomFieldValue(record, "feature_code");
  const isFacilityBooking =
    featureCode === "facility" ||
    businessValue(record, "facilitiesId", "facilityId", "fctBookingId") !== missing;
  const incomingFields = flattenFields(fieldGroups(record).requestBody);
  const items = [
    ["ประเภทเอกสาร", documentLabels[type] ?? type],
    ["การดำเนินการ", action],
    [
      "เลขบิล",
      businessValue(
        record,
        "billWatersId",
        "billCentralsId",
        "billInsurancesId",
        "billOthersId",
      ),
    ],
    ["ใบแจ้งหนี้/ใบเสร็จ", businessValue(record, "invoicesId", "receiptsId")],
    ["สมาชิก", businessValue(record, "membersId")],
    ["ห้อง/ยูนิต", businessValue(record, "addressesId")],
    ["อาคาร", businessValue(record, "buildingsId")],
    ["งวด", period],
    [
      "วันที่เอกสาร",
      businessValue(record, "askTime", "confirmTime", "dueTime", "createdAt"),
    ],
    ["ยอดเงิน", businessValue(record, "realPrice", "price", "amount")],
    ...(isFacilityBooking
      ? [
          [
            "Facility name",
            businessValue(record, "facilityName", "facility_name", "facilityId", "facilitiesId"),
          ],
          ["Booking by", businessValue(record, "actor_type", "actorType")],
          ["Booking ID", businessValue(record, "fctBookingId")],
          ["Booker name", businessValue(record, "memberName", "member_name")],
          ["Member ID", businessValue(record, "memberId", "membersId")],
          [
            "Place",
            businessValue(record, "addressName", "addressesName", "address_name"),
          ],
          ["Address ID", businessValue(record, "addressId", "addressesId")],
          ["Booking date", businessValue(record, "bookingDate")],
          ["Booking start", businessValue(record, "bookingStart")],
          ["Booking end", businessValue(record, "bookingEnd")],
          ["User amount", businessValue(record, "userAmount")],
        ]
      : []),
].filter(([, value]) => value !== missing);

  if (!items.length) {
    return (
      <Empty
        image={Empty.PRESENTED_IMAGE_SIMPLE}
        description="ไม่มีข้อมูลสำหรับสรุป"
      />
    );
  }

  return (
    <section className="log-pretty-view" aria-label="Pretty log view">
      <Descriptions
        bordered
        size="small"
        column={1}
        items={items.map(([label, value]) => ({
          key: label,
          label,
          children: value,
        }))}
      />
      {incomingFields.length > 0 && (
        <>
          <h3>ข้อมูลที่ส่งมา</h3>
          <Descriptions
            bordered
            size="small"
            column={1}
            items={incomingFields.map(([label, value]) => ({
              key: label,
              label,
              children: value,
            }))}
          />
        </>
      )}
    </section>
  );
}
