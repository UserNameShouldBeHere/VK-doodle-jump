#!/usr/bin/env tarantool

datetime = require('datetime')

box.cfg{
    listen = 3301,
    background = false
}
box.schema.user.passwd('pass')

-- ===================================

box.schema.space.create('leagues')

box.space.leagues:format({
    {name = 'id', type = 'unsigned'},
    {name = 'name', type = 'string'},
    {name = 'up_cnt', type = 'unsigned'},
    {name = 'stay_cnt', type = 'unsigned'},
})

box.space.leagues:create_index('primary', {type = 'tree', parts = {
    {'id', sort_order = 'desc'}
}})

box.space.leagues:insert{0, 'Дерево', 3, 0}
box.space.leagues:insert{1, 'Медь', 2, 3}
box.space.leagues:insert{2, 'Серебро', 1, 2}
box.space.leagues:insert{3, 'Золото', 0, 2}

box.schema.func.drop('leagues_settings', {if_exists = true})
box.schema.func.create('leagues_settings', {
    body = [[
        function()
            local res = {}
            
            for _, league in ipairs(box.space.leagues.index.primary:select({})) do
                table.insert(res, box.tuple.new({league.id, league.up_cnt, league.stay_cnt}))
            end

            return res
        end
    ]]
})

-- ===================================

box.schema.space.create('users')

box.space.users:format({
    {name = 'name', type = 'string'},
    {name = 'league', type = 'unsigned', foreign_key = {space = 'leagues', field = 'id'}},
    {name = 'max_score', type = 'unsigned'},
    {name = 'last_update', type = 'datetime'},
})

box.space.users:create_index('name', {type = 'tree', parts = {'name'}})
box.space.users:create_index('max_score', {type = 'tree', unique = false, parts = {'max_score'}})
box.space.users:create_index('score_update', {type = 'tree', parts = {
    {'max_score', sort_order = 'desc'},
    {'last_update', sort_order = 'asc'}
}})
box.space.users:create_index('league_score_update', {type = 'tree', parts = {
    'league',
    {'max_score', sort_order = 'desc'},
    {'last_update', sort_order = 'asc'}
}})

box.schema.func.drop('users_top', {if_exists = true})
box.schema.func.create('users_top', {
    body = [[
        function(limit)
            local lim = args.limit or 10
            local res = {}
            
            for _, user in ipairs(box.space.users.index.score_update:select({}, {limit = lim})) do
                table.insert(res, box.tuple.new({user.name, user.max_score}))
            end

            return res
        end
    ]]
})

box.schema.func.drop('league_users_top', {if_exists = true})
box.schema.func.create('league_users_top', {
    body = [[
        function(args)
            local leagues = {}

            for _, league in ipairs(box.space.leagues.index.primary:select({}, {iterator = box.index.GT})) do
                table.insert(leagues, box.tuple.new({league.id, league.name}))
            end
            
            local lim = args.limit or 10
            local top_users = {}
            
            for _, league in ipairs(leagues) do
                local leagueUsers = {}
                
                for _, user in ipairs(box.space.users.index.league_score_update:select({league[1]}, {limit = lim})) do
                    table.insert(leagueUsers, box.tuple.new({user.name, user.max_score}))
                end

                table.insert(top_users, box.tuple.new({league[2], leagueUsers}))
            end

            return top_users
        end
    ]]
})

box.schema.func.drop('user_score', {if_exists = true})
box.schema.func.create('user_score', {
    body = [[
        function(args)
            return box.space.users.index.name:select({args.name})[1]['max_score']
        end
    ]]
})

box.schema.func.drop('league_users_pos', {if_exists = true})
box.schema.func.create('league_users_pos', {
    body = [[
        function(args)
            local league = args.league or 1
            local res = {}
            
            for _, user in ipairs(box.space.users.index.league_score_update:select({league})) do
                table.insert(res, box.tuple.new({user.name}))
            end

            return res
        end
    ]]
})

box.schema.func.drop('league_change', {if_exists = true})
box.schema.func.create('league_change', {
    body = [[
        function(args)
            local users = args.users
            local up = args.up or false

            if (up) then
                for _, user in ipairs(users) do
                    box.space.users.index.name:update(user, {{'+', 2, 1}})
                end
            else
                for _, user in ipairs(users) do
                    box.space.users.index.name:update(user, {{'-', 2, 1}})
                end
            end
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
