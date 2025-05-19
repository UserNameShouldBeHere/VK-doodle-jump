#!/usr/bin/env tarantool

datetime = require('datetime')

box.cfg{
    listen = 3301,
    background = false
}
box.schema.user.passwd('pass')

-- ===================================

box.schema.space.create('admins')

box.space.admins:format({
    {name = 'vkid', type = 'unsigned'},
})

box.space.admins:create_index('primary', {type = 'tree', parts = {'vkid'}})

box.schema.func.drop('add_admins', {if_exists = true})
box.schema.func.create('add_admins', {
    body = [[
        function(args)
            box.space.admins:truncate({})

            for _, vkid in ipairs(args.vkids) do
                box.space.admins:insert({vkid})
            end
        end
    ]]
})

box.schema.func.drop('is_admin', {if_exists = true})
box.schema.func.create('is_admin', {
    body = [[
        function(args)
            user = box.space.admins.index.primary:select({args.vkid})

            if (user[1] == nil) then
                return false
            end

            return true
        end
    ]]
})

-- ===================================

box.schema.space.create('users')

box.space.users:format({
    {name = 'vkid', type = 'unsigned'},
    {name = 'name', type = 'string'},
    {name = 'avatar', type = 'string'},
    {name = 'access_token', type = 'string'},
    {name = 'access_update', type = 'datetime'},

    {name = 'max_score', type = 'unsigned'},
    {name = 'last_update', type = 'datetime'},
})

box.space.users:create_index('primary', {type = 'tree', parts = {'vkid'}})
box.space.users:create_index('name', {type = 'tree', parts = {'name'}})
box.space.users:create_index('max_score', {type = 'tree', unique = false, parts = {'max_score'}})
box.space.users:create_index('score_update', {type = 'tree', parts = {
    {'max_score', sort_order = 'desc'},
    {'last_update', sort_order = 'asc'}
}})

box.schema.func.drop('user_header', {if_exists = true})
box.schema.func.create('user_header', {
    body = [[
        function(args)
            user = box.space.users.index.primary:select({args.vkid})[1]

            return box.tuple.new({user.vkid, user.name, user.avatar})
        end
    ]]
})

box.schema.func.drop('has_user', {if_exists = true})
box.schema.func.create('has_user', {
    body = [[
        function(args)
            user = box.space.users.index.primary:select({args.vkid})

            if (user[1] == nil) then
                return false
            end

            return true
        end
    ]]
})

box.schema.func.drop('check_user_token', {if_exists = true})
box.schema.func.create('check_user_token', {
    body = [[
        function(args)
            user = box.space.users.index.primary:select({args.vkid})

            if (user[1].access_token ~= args.token) then
                return false
            end

            return true
        end
    ]]
})

box.schema.func.drop('check_user_access', {if_exists = true})
box.schema.func.create('check_user_access', {
    body = [[
        function(args)
            user = box.space.users.index.primary:select({args.vkid})

            if (user[1].access_token ~= args.token) then
                return false
            end

            if (user[1].access_update > datetime.now()) then
                return true
            else
                return false
            end
        end
    ]]
})

box.schema.func.drop('users_top', {if_exists = true})
box.schema.func.create('users_top', {
    body = [[
        function(args)
            local lim = args.limit or 10
            local res = {}
            
            for _, user in ipairs(box.space.users.index.score_update:select({}, {limit = lim})) do
                table.insert(res, box.tuple.new({user.name, user.max_score}))
            end

            return res
        end
    ]]
})

box.schema.func.drop('users_nearby', {if_exists = true})
box.schema.func.create('users_nearby', {
    body = [[
        function(args)
            local lim = args.limit or 10

            res = {}
            downCnt = 0
            upCnt = 0
            
            currentUser = box.space.users.index.primary:select({args.vkid}, {limit = 1})[1]

            for _, user in ipairs(box.space.users.index.score_update:select({currentUser.max_score}, {limit = lim, iterator = 'LT'})) do
                table.insert(res, box.tuple.new({user.name, user.max_score}))
                downCnt = downCnt + 1
            end

            table.insert(res, box.tuple.new({currentUser.name, currentUser.max_score}))

            for _, user in ipairs(box.space.users.index.score_update:select({currentUser.max_score}, {limit = lim, iterator = 'GT'})) do
                table.insert(res, box.tuple.new({user.name, user.max_score}))
                upCnt = upCnt + 1
            end

            return res
        end
    ]]
})

box.schema.func.drop('user_score', {if_exists = true})
box.schema.func.create('user_score', {
    body = [[
        function(args)
            return box.space.users.index.primary:select({args.vkid})[1]['max_score']
        end
    ]]
})

-- ===================================

box.schema.space.create('promocodes')
box.schema.sequence.create('promocodes_id_seq', {min = 1, start = 1})

box.space.promocodes:format({
    {name = 'id', type = 'unsigned'},
    {name = 'name', type = 'string'},
    {name = 'company', type = 'string'},
    {name = 'logo_link', type = 'string'},
    {name = 'description', type = 'string'},
    {name = 'price', type = 'unsigned'},
    {name = 'count', type = 'unsigned'},
    {name = 'code', type = 'string'},
    {name = 'activation_link', type = 'string'},
    {name = 'active_to', type = 'datetime'},
    {name = 'last_update', type = 'datetime'},
})

box.space.promocodes:create_index('primary', {sequence = 'promocodes_id_seq', type = 'tree', parts = {'id'}})
box.space.promocodes:create_index('last_update', {type = 'tree', parts = {
    {'last_update', sort_order = 'desc'}
}})

box.schema.func.drop('promocodes_for_admin', {if_exists = true})
box.schema.func.create('promocodes_for_admin', {
    body = [[
        function()
            local promocodes = {}

            for _, promocode in ipairs(box.space.promocodes.index.last_update:select({})) do
                table.insert(promocodes, box.tuple.new({
                    promocode.id,
                    promocode.name,
                    promocode.company,
                    promocode.logo_link,
                    promocode.description,
                    promocode.price,
                    promocode.count,
                    promocode.code,
                    promocode.activation_link,
                    promocode.active_to
                }))
            end

            return promocodes
        end
    ]]
})

-- ===================================

box.schema.space.create('products')
box.schema.sequence.create('products_id_seq', {min = 1, start = 1})

box.space.products:format({
    {name = 'id', type = 'unsigned'},
    {name = 'name', type = 'string'},
    {name = 'photo_link', type = 'string'},
    {name = 'description', type = 'string'},
    {name = 'price', type = 'unsigned'},
    {name = 'count', type = 'unsigned'},
    {name = 'activation_link', type = 'string'},
    {name = 'last_update', type = 'datetime'},
})

box.space.products:create_index('primary', {sequence = 'products_id_seq', type = 'tree', parts = {'id'}})
box.space.products:create_index('last_update', {type = 'tree', parts = {
    {'last_update', sort_order = 'desc'}
}})

box.schema.func.drop('products_for_admin', {if_exists = true})
box.schema.func.create('products_for_admin', {
    body = [[
        function()
            local products = {}

            for _, product in ipairs(box.space.products.index.last_update:select({})) do
                table.insert(products, box.tuple.new({
                    product.id,
                    product.name,
                    product.photo_link,
                    product.description,
                    product.price,
                    product.count,
                    product.activation_link
                }))
            end

            return products
        end
    ]]
})

-- ===================================

box.schema.space.create('tasks')
box.schema.sequence.create('tasks_id_seq', {min = 1, start = 1})

box.space.tasks:format({
    {name = 'id', type = 'unsigned'},
    {name = 'name', type = 'string'},
    {name = 'description', type = 'string'},
    {name = 'is_superpower', type = 'boolean'},
    {name = 'reward', type = 'unsigned'},
    {name = 'token', type = 'string'},
    {name = 'last_update', type = 'datetime'},
})

box.space.tasks:create_index('primary', {sequence = 'tasks_id_seq', type = 'tree', parts = {'id'}})
box.space.tasks:create_index('last_update', {type = 'tree', parts = {
    {'last_update', sort_order = 'desc'}
}})

box.schema.func.drop('tasks_for_admin', {if_exists = true})
box.schema.func.create('tasks_for_admin', {
    body = [[
        function()
            local tasks = {}

            for _, task in ipairs(box.space.tasks.index.last_update:select({})) do
                table.insert(tasks, box.tuple.new({
                    task.id,
                    task.name,
                    task.description,
                    task.reward,
                    task.token
                }))
            end

            return tasks
        end
    ]]
})

-- ===================================

box.schema.space.create('user_tasks')

box.space.user_tasks:format({
    {name = 'vkid', type = 'unsigned'},
    {name = 'task_id', type = 'unsigned'},
})

box.space.user_tasks:create_index('primary', {type = 'tree', parts = {'vkid'}})

box.schema.func.drop('tasks', {if_exists = true})
box.schema.func.create('tasks', {
    body = [[
        function(args)
            local tasks = {}

            for _, task in ipairs(box.space.tasks.index.last_update:select({})) do
                user_task = box.space.user_tasks.index.primary:select({args.vkid})

                completed = false

                if (user_task[1] ~= nil) then
                    completed = true
                end

                table.insert(tasks, box.tuple.new({
                    task.description,
                    completed
                }))
            end

            return tasks
        end
    ]]
})

-- ===================================

box.schema.space.create('shop')

box.space.shop:format({
    {name = 'vkid', type = 'unsigned'},
    {name = 'completed_tasks', type = 'array'},
})

box.space.shop:create_index('primary', {type = 'tree', parts = {'vkid'}})

box.schema.func.drop('tasks', {if_exists = true})
box.schema.func.create('tasks', {
    body = [[
        function(args)
            local tasks = {}

            return tasks
        end
    ]]
})
