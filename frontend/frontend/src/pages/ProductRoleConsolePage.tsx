import { useEffect, useMemo, useState } from 'react';
import { Navigate } from 'react-router-dom';
import { useAuthStore } from '@/store';
import { ROUTES, PERMISSIONS } from '@/constants';
import { productAdminService } from '@/services';
import { customFieldService, type CustomField } from '@/services/custom-field.service';
import type { Product, ProductRoleDefinition, RolePermissionAssignment } from '@/types';

// [KEY : ACCESS PERMISSOIN PAGE]
const RESOURCE_ACTIONS: Array<{ resource: string; actions: string[] }> = [
  { resource: 'PRODUCT', actions: ['READ', 'UPDATE', 'DELETE'] },
  { resource: 'PROJECT', actions: ['CREATE', 'READ', 'UPDATE', 'DELETE'] },
  { resource: 'FEATURE', actions: ['CREATE', 'READ', 'UPDATE', 'DELETE'] },
  { resource: 'CATEGORY', actions: ['CREATE', 'READ', 'UPDATE', 'DELETE'] },
  { resource: 'ROLE', actions: ['CREATE', 'READ', 'UPDATE', 'DELETE'] },
  { resource: 'ACCESS', actions: ['READ', 'GRANT', 'REVOKE'] },
  { resource: 'API_KEY', actions: ['CREATE', 'READ', 'UPDATE', 'DELETE'] },
  { resource: 'ENVIRONMENT', actions: ['CREATE', 'READ', 'UPDATE', 'DELETE'] },
  { resource: 'LOG', actions: ['READ', 'EXPORT', 'VIEW_SENSITIVE'] },
  { resource: 'ELASTIC_INDEX_POLICY', actions: ['CREATE', 'READ', 'UPDATE', 'DELETE'] },
];

const CUSTOM_FIELD_ACTIONS = ['VISIBLE', 'SEARCH', 'FILTER', 'SORT', 'AGGREGATE'];

function permissionKey(resourceType: string, action: string) {
  return `${resourceType}:${action}`;
}

export function ProductRoleConsolePage() {
  const permissions = useAuthStore((state) => state.permissions);
  const userRoles = useAuthStore((state) => state.roles);
  const isGodOrOwner = userRoles.includes('god') || userRoles.includes('owner') || userRoles.includes('superadmin') || userRoles.includes('super_admin');

  const canCreate = isGodOrOwner || permissions.includes(PERMISSIONS.ROLE_CREATE);
  const canEdit = isGodOrOwner || permissions.includes(PERMISSIONS.ROLE_EDIT);
  const canDelete = isGodOrOwner || permissions.includes(PERMISSIONS.ROLE_DELETE);

  if (!permissions.includes(PERMISSIONS.ROLE_VIEW) && !isGodOrOwner) {
    return <Navigate to={ROUTES.FORBIDDEN} replace />;
  }

  const [products, setProducts] = useState<Product[]>([]);
  const [roles, setRoles] = useState<ProductRoleDefinition[]>([]);
  const [customFields, setCustomFields] = useState<CustomField[]>([]);
  const [loadingCustomFields, setLoadingCustomFields] = useState(false);
  const [selectedProductId, setSelectedProductId] = useState<number | null>(null);
  const [loadingProducts, setLoadingProducts] = useState(false);
  const [loadingRoles, setLoadingRoles] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [editingRoleId, setEditingRoleId] = useState<number | null>(null);
  const [selectedPermissions, setSelectedPermissions] = useState<Record<string, boolean>>({});
  const [form, setForm] = useState({
    roleCode: '',
    roleName: '',
    isActive: true,
  });

  const selectedPermissionList = useMemo<RolePermissionAssignment[]>(() => {
    const permissions = Object.entries(selectedPermissions)
      .filter(([, enabled]) => enabled)
      .map(([key]) => {
        const separator = key.lastIndexOf(':');
        return {
          resourceType: separator >= 0 ? key.slice(0, separator) : key,
          action: separator >= 0 ? key.slice(separator + 1) : '',
        };
      });

    const existing = new Set(permissions.map((permission) => permissionKey(permission.resourceType, permission.action)));
    customFields.forEach((field) => {
      const resource = `CUSTOM_FIELD:${field.field_definition_id}`;
      const defaults: Record<string, boolean> = {
        VISIBLE: field.is_visible,
        SEARCH: field.is_searchable,
        FILTER: field.is_filterable,
        SORT: field.is_sortable,
        AGGREGATE: field.is_aggregatable,
      };
      CUSTOM_FIELD_ACTIONS.forEach((action) => {
        const key = permissionKey(resource, action);
        if (!Object.prototype.hasOwnProperty.call(selectedPermissions, key) && defaults[action]) {
          permissions.push({ resourceType: resource, action });
          existing.add(key);
        }
      });
    });
    return permissions;
  }, [selectedPermissions, customFields]);

  useEffect(() => {
    const loadProducts = async () => {
      setLoadingProducts(true);
      setError(null);
      try {
        const values = await productAdminService.listProducts();
        setProducts(values);
        if (values.length > 0) {
          setSelectedProductId(values[0].productId);
        }
      } catch (err: any) {
        setError(err?.message || 'Failed to load products');
      } finally {
        setLoadingProducts(false);
      }
    };

    void loadProducts();
  }, []);

  useEffect(() => {
    if (!selectedProductId) {
      setRoles([]);
      return;
    }

    const loadRoles = async () => {
      setLoadingRoles(true);
      setError(null);
      try {
        const values = await productAdminService.listRoles(selectedProductId);
        setRoles(values);
      } catch (err: any) {
        setError(err?.message || 'Failed to load roles');
      } finally {
        setLoadingRoles(false);
      }
    };

    void loadRoles();
  }, [selectedProductId]);

  useEffect(() => {
    if (!selectedProductId) {
      setCustomFields([]);
      return;
    }
    const loadCustomFields = async () => {
      setLoadingCustomFields(true);
      try {
        setCustomFields(await customFieldService.list(selectedProductId));
      } catch (err: any) {
        setError(err?.message || 'Failed to load custom fields');
      } finally {
        setLoadingCustomFields(false);
      }
    };
    void loadCustomFields();
  }, [selectedProductId]);

  const resetForm = () => {
    setEditingRoleId(null);
    setSelectedPermissions({});
    setForm({
      roleCode: '',
      roleName: '',
      isActive: true,
    });
  };

  const togglePermission = (resourceType: string, action: string, enabled: boolean) => {
    const key = permissionKey(resourceType, action);
    setSelectedPermissions((current) => ({ ...current, [key]: enabled }));
  };

  const handleSubmit = async (event: React.FormEvent) => {
    event.preventDefault();
    if (!selectedProductId) {
      setError('Select a product first');
      return;
    }
    if (selectedPermissionList.length === 0) {
      setError('Select at least one permission');
      return;
    }

    setSubmitting(true);
    setError(null);
    try {
      if (editingRoleId) {
        const updated = await productAdminService.updateRole(selectedProductId, editingRoleId, {
          roleName: form.roleName,
          permissions: selectedPermissionList,
          isActive: form.isActive,
        });
        setRoles((current) => current.map((item) => (item.roleId === updated.roleId ? updated : item)));
      } else {
        const created = await productAdminService.createRole(selectedProductId, {
          roleCode: form.roleCode,
          roleName: form.roleName,
          permissions: selectedPermissionList,
        });
        setRoles((current) => [...current, created]);
      }
      resetForm();
    } catch (err: any) {
      setError(err?.message || 'Failed to save role');
    } finally {
      setSubmitting(false);
    }
  };

  const handleEdit = (role: ProductRoleDefinition) => {
    const mapped: Record<string, boolean> = {};
    role.permissions.forEach((permission) => {
      mapped[permissionKey(permission.resourceType, permission.action)] = true;
    });
    setEditingRoleId(role.roleId);
    setSelectedPermissions(mapped);
    setForm({
      roleCode: role.roleCode,
      roleName: role.roleName,
      isActive: role.isActive,
    });
  };

  const handleDelete = async (role: ProductRoleDefinition) => {
    if (!selectedProductId) {
      return;
    }
    if (!window.confirm(`Delete role "${role.roleName}"?`)) {
      return;
    }

    setSubmitting(true);
    setError(null);
    try {
      await productAdminService.deleteRole(selectedProductId, role.roleId);
      setRoles((current) => current.filter((item) => item.roleId !== role.roleId));
      if (editingRoleId === role.roleId) {
        resetForm();
      }
    } catch (err: any) {
      setError(err?.message || 'Failed to delete role');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="p-6 max-w-7xl mx-auto">
      <h1 className="text-xl font-semibold mb-4">Product Role Console</h1>

      <div className="mb-4">
        <label className="block text-sm mb-1">Product</label>
        <select
          value={selectedProductId ?? ''}
          onChange={(event) => setSelectedProductId(Number(event.target.value) || null)}
          className="border px-3 py-2 w-full max-w-md"
          disabled={loadingProducts}
        >
          <option value="">Select product</option>
          {products.map((product) => (
            <option key={product.productId} value={product.productId}>
              {product.productName} ({product.productCode})
            </option>
          ))}
        </select>
      </div>

      {error ? <div className="mb-4 border border-red-300 bg-red-50 p-3 text-sm text-red-700">{error}</div> : null}

      <div className="grid gap-6 lg:grid-cols-[1.3fr_1fr]">
        {(canCreate && !editingRoleId) || (canEdit && editingRoleId) ? (
          <section className="border p-4">
          <h2 className="font-medium mb-3">{editingRoleId ? 'Edit role' : 'Create role'}</h2>
          <form onSubmit={handleSubmit} className="space-y-4">
            <div className="grid gap-3 md:grid-cols-2">
              <div>
                <label className="block text-sm mb-1">Role code</label>
                <input
                  type="text"
                  value={form.roleCode}
                  onChange={(event) => setForm((current) => ({ ...current, roleCode: event.target.value }))}
                  className="border px-3 py-2 w-full"
                  disabled={Boolean(editingRoleId)}
                  required
                />
              </div>
              <div>
                <label className="block text-sm mb-1">Role name</label>
                <input
                  type="text"
                  value={form.roleName}
                  onChange={(event) => setForm((current) => ({ ...current, roleName: event.target.value }))}
                  className="border px-3 py-2 w-full"
                  required
                />
              </div>
            </div>

            {editingRoleId ? (
              <label className="flex items-center gap-2 text-sm">
                <input
                  type="checkbox"
                  checked={form.isActive}
                  onChange={(event) => setForm((current) => ({ ...current, isActive: event.target.checked }))}
                />
                Active
              </label>
            ) : null}

            <div>
              <div className="text-sm font-medium mb-2">Permissions</div>
              <div className="overflow-x-auto">
                <table className="min-w-full border text-sm">
                  <thead>
                    <tr className="bg-gray-50">
                      <th className="border px-3 py-2 text-left">Resource</th>
                      <th className="border px-3 py-2 text-left">Actions</th>
                    </tr>
                  </thead>
                  <tbody>
                    {RESOURCE_ACTIONS.map((item) => (
                      <tr key={item.resource}>
                        <td className="border px-3 py-2 align-top font-medium">{item.resource}</td>
                        <td className="border px-3 py-2">
                          <div className="flex flex-wrap gap-3">
                            {item.actions.map((action) => (
                              <label key={action} className="flex items-center gap-2">
                                <input
                                  type="checkbox"
                                  checked={Boolean(selectedPermissions[permissionKey(item.resource, action)])}
                                  onChange={(event) => togglePermission(item.resource, action, event.target.checked)}
                                />
                                <span>{action}</span>
                              </label>
                            ))}
                          </div>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>

            <div className="mt-5">
              <div className="text-sm font-medium mb-2">Custom Field capabilities</div>
              <div className="text-xs text-gray-500 mb-2">
                ตั้งค่าสิทธิ์การแสดงผลและการใช้งาน Custom Field สำหรับ Role นี้โดยตรง
              </div>
              {loadingCustomFields ? <div className="text-sm text-gray-500">Loading custom fields...</div> : null}
              {!loadingCustomFields && customFields.length === 0 ? (
                <div className="text-sm text-gray-500 border p-3">ยังไม่มี Custom Field ใน Product นี้</div>
              ) : null}
              {!loadingCustomFields && customFields.length > 0 ? (
                <div className="overflow-x-auto">
                  <table className="min-w-full border text-sm">
                    <thead>
                      <tr className="bg-gray-50">
                        <th className="border px-3 py-2 text-left">Field</th>
                        {CUSTOM_FIELD_ACTIONS.map((action) => (
                          <th key={action} className="border px-3 py-2 text-center">{action}</th>
                        ))}
                      </tr>
                    </thead>
                    <tbody>
                      {customFields.map((field) => {
                        const resource = `CUSTOM_FIELD:${field.field_definition_id}`;
                        const defaultValues: Record<string, boolean> = {
                          VISIBLE: field.is_visible,
                          SEARCH: field.is_searchable,
                          FILTER: field.is_filterable,
                          SORT: field.is_sortable,
                          AGGREGATE: field.is_aggregatable,
                        };
                        return (
                          <tr key={field.field_definition_id}>
                            <td className="border px-3 py-2">
                              <div className="font-medium">{field.display_name || field.field_key}</div>
                              <div className="text-xs text-gray-500">{field.field_path || field.field_key}</div>
                            </td>
                            {CUSTOM_FIELD_ACTIONS.map((action) => {
                              const key = permissionKey(resource, action);
                              const checked = Object.prototype.hasOwnProperty.call(selectedPermissions, key)
                                ? Boolean(selectedPermissions[key])
                                : defaultValues[action];
                              return (
                                <td key={action} className="border px-3 py-2 text-center">
                                  <input
                                    type="checkbox"
                                    checked={checked}
                                    onChange={(event) => togglePermission(resource, action, event.target.checked)}
                                  />
                                </td>
                              );
                            })}
                          </tr>
                        );
                      })}
                    </tbody>
                  </table>
                </div>
              ) : null}
            </div>

            <div className="text-xs text-gray-500">
              Selected: {selectedPermissionList.length} permission{selectedPermissionList.length === 1 ? '' : 's'}
            </div>

            <div className="flex gap-2">
              <button type="submit" className="border px-3 py-2" disabled={submitting || !selectedProductId}>
                {submitting ? 'Saving...' : editingRoleId ? 'Update role' : 'Create role'}
              </button>
              <button type="button" className="border px-3 py-2" onClick={resetForm}>
                Clear
              </button>
            </div>
          </form>
        </section>
        ) : (
          <div className="hidden lg:block"></div>
        )}

        <section className="border p-4">
          <h2 className="font-medium mb-3">Roles</h2>
          {loadingRoles ? <div className="text-sm">Loading roles...</div> : null}
          {!loadingRoles && roles.length === 0 ? <div className="text-sm text-gray-500">No roles yet.</div> : null}
          <div className="space-y-3">
            {roles.map((role) => (
              <div key={role.roleId} className="border p-3">
                <div className="flex items-start justify-between gap-3">
                  <div>
                    <div className="font-medium">{role.roleName}</div>
                    <div className="text-sm text-gray-600">{role.roleCode}</div>
                    <div className="text-xs text-gray-500 mt-1">Status: {role.isActive ? 'active' : 'inactive'}</div>
                  </div>
                  <div className="flex gap-2">
                    {canEdit && (
                      <button type="button" className="border px-2 py-1 text-sm" onClick={() => handleEdit(role)}>
                        Edit
                      </button>
                    )}
                    {canDelete && (
                      <button type="button" className="border px-2 py-1 text-sm" onClick={() => handleDelete(role)}>
                        Delete
                      </button>
                    )}
                  </div>
                </div>
                <div className="mt-3 flex flex-wrap gap-2">
                  {role.permissions.map((permission) => (
                    <span key={permissionKey(permission.resourceType, permission.action)} className="border px-2 py-1 text-xs">
                      {permission.resourceType}.{permission.action}
                    </span>
                  ))}
                </div>
              </div>
            ))}
          </div>
        </section>
      </div>
    </div>
  );
}
